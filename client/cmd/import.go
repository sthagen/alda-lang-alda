package cmd

import (
	"alda.io/client/code-generator"
	"alda.io/client/help"
	"alda.io/client/interoperability"
	"alda.io/client/model"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var outputAldaFilename string
var importFormat string

func init() {
	importCmd.Flags().StringVarP(
		&file, "file", "f", "", "Read data from a file to convert to Alda",
	)

	importCmd.Flags().StringVarP(
		&code, "code", "c", "", "Read data from a string to convert to Alda",
	)

	importCmd.Flags().StringVarP(
		&outputAldaFilename, "output", "o", "", "The output Alda code filename",
	)

	importCmd.Flags().StringVarP(
		&importFormat, "import-format", "i", "", "The format of the imported data",
	)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Evaluate external format and import as Alda source code",
	Long:  `Evaluate external format and import as Alda source code

---

Source code can be provided in one of three ways:

The path to a file (-f, --file):
  alda import -i musicxml -f path/to/my-score.musicxml -o my-score.alda

A string of code (-c, --code):
  alda import -i musicxml -c "...some musicxml data..." -o my-score.alda

Text piped into the process on stdin:
  echo "...some musicxml data..." | alda import -i musicxml -o my-score.alda

---

When -o / --output FILENAME is provided, the results are written into that file.

  alda import -i musicxml -c "...some musicxml data..." -o my-score.alda

Otherwise, the results are written to stdout, which is convenient for
redirecting into other files or processes.

  alda import -i musicxml -f path/to/my-score.musicxml > my-score.alda
  alda import -i musicxml -f path/to/my-score.musicxml | some-process > my-score.alda

---

Currently, the only import format is MusicXML.  

---`,
	RunE:  func(_ *cobra.Command, args[]string) error {
		if importFormat != "musicxml" {
			return help.UserFacingErrorf(
				`%s is not a supported input format.

Currently, the only supported output format is %s.`,
				aurora.BrightYellow(importFormat),
				aurora.BrightYellow("musicxml"),
			)
		}

		var score *model.Score
		var err error

		// TODO: add XML validation
		// TODO: add XML conversion to ensure we get score-partwise pieces as input
		switch {
		case file != "":
			inputFile, err := os.Open(file)
			if err != nil {
				return err
			}

			score, err = interoperability.ImportMusicXML(inputFile)
		case code != "":
			reader := strings.NewReader(code)
			score, err = interoperability.ImportMusicXML(reader)

		default:
			bytes, err := readStdin()
			if err != nil {
				return err
			}

			reader := strings.NewReader(string(bytes))
			score, err = interoperability.ImportMusicXML(reader)
		}

		if err != nil {
			return err
		}

		if outputAldaFilename == "" {
			// When no output filename is specified, we write directly to stdout
			code_generator.Generate(score, os.Stdout)
			return nil
		} else {
			file, err := os.Create(outputAldaFilename)
			if err != nil {
				return err
			}

			code_generator.Generate(score, file)

			fmt.Fprintf(os.Stderr, "Imported score to %s\n", outputAldaFilename)
			if err := file.Close(); err != nil {
				return err
			}
			return nil
		}
	},
}
