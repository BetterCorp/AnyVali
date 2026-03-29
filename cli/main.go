package main

import (
	"fmt"
	"os"

	"github.com/anyvali/anyvali/cli/cmd"
)

const usage = `Usage: anyvali <command> [options]

Commands:
  validate    Validate input against a schema
  serve       Start HTTP validation server
  inspect     Show schema structure
  check       Check if a schema is portable
  version     Print version

Flags:
  -h, --help    Show help
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	command := os.Args[1]

	switch command {
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	case "validate":
		os.Exit(cmd.RunValidate(os.Args[2:]))
	case "serve":
		os.Exit(cmd.RunServe(os.Args[2:]))
	case "inspect":
		os.Exit(cmd.RunInspect(os.Args[2:]))
	case "check":
		os.Exit(cmd.RunCheck(os.Args[2:]))
	case "version":
		os.Exit(cmd.RunVersion(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
