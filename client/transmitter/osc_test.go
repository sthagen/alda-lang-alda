package transmitter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/daveyarwood/go-osc/osc"

	"alda.io/client/model"
	"alda.io/client/parser"
	_ "alda.io/client/testing"
)

// parseScore builds a score by parsing the provided Alda code.
func parseScore(t *testing.T, input string) *model.Score {
	t.Helper()

	ast, err := parser.ParseString(input)
	if err != nil {
		t.Fatal(err)
	}

	updates, err := ast.Updates()
	if err != nil {
		t.Fatal(err)
	}

	score := model.NewScore()
	if err := score.Update(updates...); err != nil {
		t.Fatal(err)
	}

	return score
}

// noteOffsetsInBundle returns the offset argument of each /track/{n}/midi/note
// message in the bundle, in the order the messages appear.
func noteOffsetsInBundle(bundle *osc.Bundle) []float64 {
	offsets := []float64{}
	for _, message := range bundle.Messages {
		if strings.HasSuffix(message.Address, "/midi/note") {
			offset, ok := message.Arguments[1].(int32)
			if !ok {
				panic(fmt.Sprintf("expected int32 note offset, got %#v", message.Arguments[1]))
			}
			offsets = append(offsets, float64(offset))
		}
	}
	return offsets
}

func expectNoteOffsets(t *testing.T, label string, actual []float64, expected ...float64) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("%s: expected note offsets %v, got %v", label, expected, actual)
	}

	for i, e := range expected {
		if actual[i] != e {
			t.Fatalf("%s: expected note offset %v, got %v", label, expected, actual)
		}
	}
}

func TestScoreToOSCBundleSyncOffsets(t *testing.T) {
	// The full score consists of two identical lines, each ending in a trailing
	// r4. The 8 notes of the first line are at offsets 0, 500, ..., 3500 (tempo
	// 240, 250ms per quarter note). The trailing r4 advances the score's offset
	// to 4000, but produces no event, so the player's position after the first
	// line is the end of the last note: 3500 + 250 = 3750.
	//
	// The 8 notes of the second line are at offsets 4000, 4500, ..., 7500. When
	// transmitted with a sync offset of 3750 (the player's position), they should
	// be re-based to 250, 750, ..., 3750, preserving the quarter-note gap
	// created by the trailing rest.
	score := parseScore(
		t,
		"percussion: o2 (tempo 240) c4 r4 c4 r4 | c4 r4 c4 r4 | c4 r4 c4 r4 | c4 r4 d4 r4\n"+
			"percussion: o2 (tempo 240) c4 r4 c4 r4 | c4 r4 c4 r4 | c4 r4 c4 r4 | c4 r4 d4 r4",
	)

	if len(score.Events) != 16 {
		t.Fatalf("expected 16 events, got %d", len(score.Events))
	}

	// First line: no sync offset, notes at their actual offsets.
	bundle, err := (OSCTransmitter{}).ScoreToOSCBundle(
		score, TransmitFromIndex(0), TransmitToIndex(8),
	)
	if err != nil {
		t.Fatal(err)
	}
	expectNoteOffsets(
		t, "first line",
		noteOffsetsInBundle(bundle),
		0, 500, 1000, 1500, 2000, 2500, 3000, 3500,
	)

	// Second line: re-based relative to the player's position (3750), so that
	// the first new note lands a quarter note after the previous line's last
	// note instead of immediately after it.
	partID := score.Events[0].(model.NoteEvent).Part.ID
	bundle, err = (OSCTransmitter{}).ScoreToOSCBundle(
		score,
		TransmitFromIndex(8),
		SyncOffsets(map[string]float64{partID: 3750}),
	)
	if err != nil {
		t.Fatal(err)
	}
	expectNoteOffsets(
		t, "second line",
		noteOffsetsInBundle(bundle),
		250, 750, 1250, 1750, 2250, 2750, 3250, 3750,
	)
}

func TestScoreToOSCBundleSyncOffsetsClampsNegativeOffsets(t *testing.T) {
	// A note whose actual offset is behind the sync offset would be scheduled
	// at a negative relative offset, which the player cannot handle. It must be
	// clamped to 0, while notes at/after the sync offset keep their spacing.
	score := parseScore(t, "piano: c1 c4")

	partID := score.Events[0].(model.NoteEvent).Part.ID
	bundle, err := (OSCTransmitter{}).ScoreToOSCBundle(
		score, SyncOffsets(map[string]float64{partID: 1000}),
	)
	if err != nil {
		t.Fatal(err)
	}

	expectNoteOffsets(t, "clamped note offsets", noteOffsetsInBundle(bundle), 0, 1000)
}

func TestScoreToOSCBundleEmptySyncOffsets(t *testing.T) {
	// An empty sync offsets map means "no synchronization"; note offsets are
	// transmitted as-is.
	score := parseScore(t, "piano: c4 d4 e4 f4")

	bundle, err := (OSCTransmitter{}).ScoreToOSCBundle(
		score, SyncOffsets(map[string]float64{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	expectNoteOffsets(t, "no sync", noteOffsetsInBundle(bundle), 0, 500, 1000, 1500)
}
