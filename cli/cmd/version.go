package cmd

import "fmt"

// Version is the CLI version string.
const Version = "0.0.1"

// RunVersion prints the version and returns exit code.
func RunVersion(args []string) int {
	fmt.Printf("anyvali version %s\n", Version)
	return 0
}
