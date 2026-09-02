package parser

import "alda.io/client/model"

func init() {
	model.ParseAldaCode = func(code string) ([]model.ScoreUpdate, error) {
		ast, err := ParseString(code)
		if err != nil {
			return nil, err
		}
		return ast.Updates()
	}
}
