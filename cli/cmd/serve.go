package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/BetterCorp/AnyVali/cli/server"
)

const serveUsage = `Usage: anyvali serve [options]

Start an HTTP validation server.

Flags:
  -p, --port     Port number (default 8080)
      --host     Bind address (default "0.0.0.0")
      --schemas  Directory to load schema files from (optional)
      --cors     Enable CORS headers (default false)
  -h, --help     Show help
`

// RunServe executes the serve command and returns the exit code.
func RunServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var port int
	var host string
	var schemasDir string
	var cors bool
	var help bool

	fs.IntVar(&port, "p", 8080, "")
	fs.IntVar(&port, "port", 8080, "")
	fs.StringVar(&host, "host", "0.0.0.0", "")
	fs.StringVar(&schemasDir, "schemas", "", "")
	fs.BoolVar(&cors, "cors", false, "")
	fs.BoolVar(&help, "h", false, "")
	fs.BoolVar(&help, "help", false, "")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 2
	}

	if help {
		fmt.Print(serveUsage)
		return 0
	}

	cfg := server.Config{
		Host:       host,
		Port:       port,
		SchemasDir: schemasDir,
		CORS:       cors,
	}

	srv, err := server.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 2
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 2
		}
	case <-stop:
		fmt.Println("\nShutting down...")
		srv.Shutdown()
	}

	return 0
}
