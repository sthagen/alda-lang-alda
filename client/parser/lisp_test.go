package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
	"github.com/go-test/deep"
)

func lispSymbol(name string) model.LispSymbol {
	return model.LispSymbol{Name: name}
}

func lispNumber(value float64) model.LispNumber {
	return model.LispNumber{Value: value}
}

func lispString(value string) model.LispString {
	return model.LispString{Value: value}
}

func lispList(elements ...model.LispForm) model.LispList {
	return model.LispList{Elements: elements}
}

func lispQuotedForm(form model.LispForm) model.LispQuotedForm {
	return model.LispQuotedForm{Form: form}
}

func lispQuotedList(elements ...model.LispForm) model.LispQuotedForm {
	return lispQuotedForm(lispList(elements...))
}

func TestLisp(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "attribute change with no value",
			given: "(fff)",
			expectUpdates: []model.ScoreUpdate{
				lispList(lispSymbol("fff")),
			},
		},
		parseTestCase{
			label: "attribute change with number value",
			given: "(volume 50)",
			expectUpdates: []model.ScoreUpdate{
				lispList(lispSymbol("volume"), lispNumber(50)),
			},
		},
		parseTestCase{
			label: "attribute change with string value",
			given: `(key-signature "f+ c+ g+")`,
			expectUpdates: []model.ScoreUpdate{
				lispList(lispSymbol("key-signature"), lispString("f+ c+ g+")),
			},
		},
		parseTestCase{
			label: "global attribute change",
			given: "(tempo! 200)",
			expectUpdates: []model.ScoreUpdate{
				lispList(lispSymbol("tempo!"), lispNumber(200)),
			},
		},
		parseTestCase{
			label: "attribute change with quoted list argument",
			given: "(key-sig '(a major))",
			expectUpdates: []model.ScoreUpdate{
				lispList(
					lispSymbol("key-sig"),
					lispQuotedList(lispSymbol("a"), lispSymbol("major")),
				),
			},
		},
		parseTestCase{
			label: "attribute change with quoted nested list argument",
			given: "(key-signature '(e (flat) b (flat)))",
			expectUpdates: []model.ScoreUpdate{
				lispList(
					lispSymbol("key-signature"),
					lispQuotedList(
						lispSymbol("e"),
						lispList(lispSymbol("flat")),
						lispSymbol("b"),
						lispList(lispSymbol("flat")),
					),
				),
			},
		},
		parseTestCase{
			label: "alda-code with plain notes",
			given: `(alda-code "piano: c d e")`,
			expectUpdates: []model.ScoreUpdate{
				lispList(lispSymbol("alda-code"), lispString("piano: c d e")),
			},
		},
	)
}

func TestAldaCodeEquivalence(t *testing.T) {
	astFromAldaCode, err := Parse(
		"alda-code", `(alda-code "piano: c d e")`, SuppressSourceContext,
	)
	if err != nil {
		t.Fatal(err)
	}
	updatesFromAldaCode, err := astFromAldaCode.Updates()
	if err != nil {
		t.Fatal(err)
	}
	scoreFromAldaCode := model.NewScore()
	if err := scoreFromAldaCode.Update(updatesFromAldaCode...); err != nil {
		t.Fatal(err)
	}

	astDirect, err := Parse(
		"direct", "piano: c d e", SuppressSourceContext,
	)
	if err != nil {
		t.Fatal(err)
	}
	updatesDirect, err := astDirect.Updates()
	if err != nil {
		t.Fatal(err)
	}
	scoreDirect := model.NewScore()
	if err := scoreDirect.Update(updatesDirect...); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(scoreDirect.Parts, scoreFromAldaCode.Parts); diff != nil {
		t.Error("Parts differ between alda-code and direct notation:")
		for _, d := range diff {
			t.Error(d)
		}
	}

	if diff := deep.Equal(scoreDirect.Events, scoreFromAldaCode.Events); diff != nil {
		t.Error("Events differ between alda-code and direct notation:")
		for _, d := range diff {
			t.Error(d)
		}
	}
}
