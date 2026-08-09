package repl

import (
	"fmt"
	"strings"
	"testing"

	"github.com/daveyarwood/go-osc/osc"

	"alda.io/client/model"
	"alda.io/client/transmitter"
	_ "alda.io/client/testing"
)

// newTestServer returns a REPL server with a fresh score and no other state,
// which is all that updateScoreWithInput requires.
func newTestServer() *Server {
	return &Server{score: model.NewScore()}
}

// updateAndBuildBundle applies the provided input to the server and returns the
// OSC bundle that would be transmitted for just the new events.
func updateAndBuildBundle(t *testing.T, server *Server, input string) *osc.Bundle {
	t.Helper()

	opts, err := server.updateScoreWithInput(input)
	if err != nil {
		t.Fatal(err)
	}

	bundle, err := (transmitter.OSCTransmitter{}).ScoreToOSCBundle(
		server.score, opts...,
	)
	if err != nil {
		t.Fatal(err)
	}

	return bundle
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
			t.Fatalf("%s: expected note offsets %v, got %v", label, expected, actual)
		}
	}
}

// When the same line is entered twice, the trailing rest at the end of the
// first line should not be stripped — the first note of the second line should
// land a quarter note late, not immediately after the last note of the first
// line.
//
// At tempo 240 (250ms per quarter note), the first line's notes are at offsets
// 0, 500, ..., 3500, and the trailing r4 advances the score's offset to 4000
// without producing an event. The player's position after the first line is the
// end of the last note (3500 + 250 = 3750). The second line's notes (at
// absolute offsets 4000..7500) must therefore be re-based relative to 3750,
// i.e. 250, 750, ..., 3750.
func TestUpdateScoreWithInputTrailingRest(t *testing.T) {
	server := newTestServer()
	line := "percussion: o2 (tempo 240) c4 r4 c4 r4 | c4 r4 c4 r4 | c4 r4 c4 r4 | c4 r4 d4 r4"

	bundle1 := updateAndBuildBundle(t, server, line)
	expectNoteOffsets(
		t, "first line",
		noteOffsetsInBundle(bundle1),
		0, 500, 1000, 1500, 2000, 2500, 3000, 3500,
	)

	bundle2 := updateAndBuildBundle(t, server, line)
	expectNoteOffsets(
		t, "second line",
		noteOffsetsInBundle(bundle2),
		250, 750, 1250, 1750, 2250, 2750, 3250, 3750,
	)
}

// When the previous line doesn't end in a rest, the player's position equals
// the earliest offset of the new events, so the re-basing is a no-op and
// behavior is unchanged.
func TestUpdateScoreWithInputNoTrailingRest(t *testing.T) {
	server := newTestServer()

	updateAndBuildBundle(t, server, "piano: c4 d4 e4 f4")

	bundle := updateAndBuildBundle(t, server, "piano: g4 a4 b4 > c4")
	expectNoteOffsets(t, "second line", noteOffsetsInBundle(bundle), 0, 500, 1000, 1500)
}

// A line that consists only of rests produces no events, but the rests must
// still delay subsequent notes. Because the sync offset is the player's
// position (the end of the last transmitted note), and the next note's absolute
// offset already includes the rests, the gap is preserved in the relative
// offset.
func TestUpdateScoreWithInputRestOnlyLine(t *testing.T) {
	server := newTestServer()

	updateAndBuildBundle(t, server, "piano: c4 d4 e4 f4")

	bundle := updateAndBuildBundle(t, server, "piano: r4 r4")
	if len(noteOffsetsInBundle(bundle)) != 0 {
		t.Fatalf("expected no notes in rest-only bundle, got %v", noteOffsetsInBundle(bundle))
	}

	// c4 d4 e4 f4 at tempo 120 = notes at 0, 500, 1000, 1500; r4 r4 advances
	// the offset to 3000. The g4 is at absolute offset 3000, and the player's
	// position is 2000, so the relative offset should be 1000.
	bundle = updateAndBuildBundle(t, server, "piano: g4")
	expectNoteOffsets(t, "note after rests", noteOffsetsInBundle(bundle), 1000)
}

// Voices should play immediately on subsequent evaluations, without
// accumulating delay (issue #415).
func TestUpdateScoreWithInputVoicesPlayImmediately(t *testing.T) {
	server := newTestServer()
	line := "piano: V1: c d e V2: c d e"

	bundle1 := updateAndBuildBundle(t, server, line)
	expectNoteOffsets(
		t, "first line",
		noteOffsetsInBundle(bundle1),
		0, 0, 500, 500, 1000, 1000,
	)

	bundle2 := updateAndBuildBundle(t, server, line)
	expectNoteOffsets(
		t, "second line",
		noteOffsetsInBundle(bundle2),
		0, 0, 500, 500, 1000, 1000,
	)
}

// When new notes are added to a voice that is behind the player's position,
// they must not be scheduled at negative relative offsets. The sync offset is
// clamped to the earliest new offset, so the earliest new note plays right away
// and the spacing between the new notes is preserved.
func TestUpdateScoreWithInputVoiceBehindPlayer(t *testing.T) {
	server := newTestServer()

	// V1 ends at offset 2000; V2 (a half note) ends at offset 1000. The
	// player's position on this track is therefore 2000.
	updateAndBuildBundle(t, server, "piano: V1: c d e f | V2: c2")

	// V2 continues at offset 1000, which is behind the player's position of
	// 2000. The two new notes at offsets 1000 and 2000 are re-based relative to
	// 1000 (the clamped sync offset), so they're transmitted as 0 and 1000.
	bundle := updateAndBuildBundle(t, server, "V2: d2 e2")
	expectNoteOffsets(t, "voice behind player", noteOffsetsInBundle(bundle), 0, 1000)
}

// A single global sync offset is an approximation when tracks sit at different
// positions. The per-part sync offset keeps each track correct: here, piano
// continues at offset 2000 while a new violin part starts at offset 0.
func TestUpdateScoreWithInputTwoTracksAtDifferentPositions(t *testing.T) {
	server := newTestServer()

	updateAndBuildBundle(t, server, "piano: c4 d4 e4 f4")

	bundle := updateAndBuildBundle(
		t,
		server,
		"piano: g4 a4 b4 > c4 | violin: c4 d4 e4 f4",
	)
	// The bundle is sorted by each event's absolute offset before the sync
	// offsets are subtracted, so the violin notes (absolute 0..1500) come
	// before the piano notes (absolute 2000..3500). Each is re-based relative
	// to its own track's position, so both come out at 0..1500.
	expectNoteOffsets(
		t, "two tracks",
		noteOffsetsInBundle(bundle),
		0, 500, 1000, 1500, 0, 500, 1000, 1500,
	)
}
