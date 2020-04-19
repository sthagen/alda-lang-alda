package cmd

import (
	"fmt"
	"os"

	log "alda.io/client/logging"
	"github.com/spf13/cobra"
)

var verbosity int

func init() {
	rootCmd.PersistentFlags().IntVarP(
		&verbosity, "verbosity", "v", 1, "verbosity level (0-3)",
	)

	for _, cmd := range []*cobra.Command{
		doctorCmd,
		playCmd,
		psCmd,
		versionCmd,
	} {
		rootCmd.AddCommand(cmd)
	}
}

func handleVerbosity(level int) error {
	switch level {
	case 0:
		log.SetGlobalLevel("error")
	case 1:
		log.SetGlobalLevel("warn")
	case 2:
		log.SetGlobalLevel("info")
	case 3:
		log.SetGlobalLevel("debug")
	default:
		return fmt.Errorf("invalid verbosity level. Valid values are 0-3")
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "alda",
	Short: "alda: a text-based language for music composition",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := handleVerbosity(verbosity); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// Execute parses command-line arguments and runs the Alda command-line client.
func Execute() error {
	// The default logging level is WARN.
	level := 1

	// Cobra is a little bit janky in that there doesn't seem to be a way to add a
	// hook that will _always_ run regardless of the command. PersistentPreRun
	// almost works, but it doesn't run in cases like:
	//
	//   alda
	//   alda --help
	//   alda some-bogus-command
	//
	// We always want to do two things:
	//
	//   1. Set the logging level based on the -v / --verbosity option.
	//   2. Spawn player processes.
	//
	// It's especially important that we do this when the user first installs Alda
	// and they run `alda` or `alda --help` to explore the commands and options.
	// By the time they figure out how to craft an `alda play ...` command, a
	// player process will have fully spawned in the background and they'll hear
	// output immediately, which is a great user experience.
	//
	// As a hacky workaround, we scan through the command line arguments here,
	// make an attempt to set the log level, and spawn player processes, before we
	// invoke Cobra and handle CLI arguments properly.
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "-v0", "-v=0", "--verbosity=0":
			level = 0
		case "-v1", "-v=1", "--verbosity=1":
			level = 1
		case "-v2", "-v=2", "--verbosity=2":
			level = 2
		case "-v3", "-v=3", "--verbosity=3":
			level = 3
		case "-v", "--verbosity":
			i++
			switch os.Args[i] {
			case "0":
				level = 0
			case "1":
				level = 1
			case "2":
				level = 2
			case "3":
				level = 3
			}
		}

		if err := handleVerbosity(level); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Regardless of the command being run*, Alda will preemptively spawn player
	// processes in the background, up to a desired amount. This helps to ensure
	// that the application will feel fast, because each time you need a player
	// process, there will probably already be one available.
	//
	// *`alda ps` is an exception because it is designed to be run repeatedly,
	// e.g. `watch -n 0.25 alda ps`, in order to provide a live-updating view of
	// current Alda processes.
	commandIsProbablyPs := false

	for _, arg := range os.Args {
		if arg == "ps" {
			commandIsProbablyPs = true
		}
	}

	if !commandIsProbablyPs {
		if err := fillPlayerPool(); err != nil {
			log.Warn().Err(err).Msg("Failed to fill player pool.")
		}
	}

	return rootCmd.Execute()
}