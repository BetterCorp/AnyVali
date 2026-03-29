package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	anyvali "github.com/anyvali/anyvali/sdk/go"
)

const validateUsage = `Usage: anyvali validate [options] <schema-file> [input]

Arguments:
  schema-file    Path to AnyVali JSON schema document
  input          JSON string, file path, or "-" for stdin

Flags:
  -f, --format   Output format: json, text, quiet (default "json")
  -s, --strict   Exit code 1 on any issue (default behavior)
      --no-color  Disable colored output
  -h, --help     Show help
`

// RunValidate executes the validate command and returns the exit code.
func RunValidate(args []string) int {
	var format string
	var noColor bool
	var help bool

	// Manual arg parsing to support both flags and positional args in any order
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			help = true
		case "-s", "--strict":
			// strict is default, just consume
		case "--no-color":
			noColor = true
		case "-f", "--format":
			if i+1 < len(args) {
				i++
				format = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a value\n", arg)
				return 2
			}
		default:
			if strings.HasPrefix(arg, "-f=") {
				format = strings.TrimPrefix(arg, "-f=")
			} else if strings.HasPrefix(arg, "--format=") {
				format = strings.TrimPrefix(arg, "--format=")
			} else if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Error: unknown flag %s\n", arg)
				return 2
			} else {
				positional = append(positional, arg)
			}
		}
	}

	if format == "" {
		format = "json"
	}

	if help {
		fmt.Print(validateUsage)
		return 0
	}

	if len(positional) < 1 {
		fmt.Fprint(os.Stderr, "Error: schema-file is required\n\n")
		fmt.Fprint(os.Stderr, validateUsage)
		return 2
	}

	schemaFile := positional[0]

	// Load schema
	schemaData, err := os.ReadFile(schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read schema file: %s\n", err)
		return 2
	}

	schema, err := anyvali.ImportJSON(schemaData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid schema: %s\n", err)
		return 2
	}

	// Read input
	var inputData []byte
	if len(positional) >= 2 {
		inputSource := positional[1]
		if inputSource == "-" {
			inputData, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: cannot read stdin: %s\n", err)
				return 2
			}
		} else if fileExists(inputSource) {
			inputData, err = os.ReadFile(inputSource)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: cannot read input file: %s\n", err)
				return 2
			}
		} else {
			// Treat as raw JSON string
			inputData = []byte(inputSource)
		}
	} else {
		// Read from stdin
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot read stdin: %s\n", err)
			return 2
		}
	}

	// Parse input JSON
	var input any
	if err := json.Unmarshal(inputData, &input); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid JSON input: %s\n", err)
		return 2
	}

	// Validate
	result := schema.SafeParse(input)

	// Output
	switch format {
	case "json":
		return outputJSON(result)
	case "text":
		return outputText(result, noColor)
	case "quiet":
		if result.Success {
			return 0
		}
		return 1
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format %q\n", format)
		return 2
	}
}

// ValidateInput is an exported helper for testing that validates input against a schema
// and returns the result in the specified format.
func ValidateInput(schema anyvali.Schema, input any, format string, noColor bool) (string, int) {
	result := schema.SafeParse(input)

	switch format {
	case "json":
		output := formatJSON(result)
		code := 0
		if !result.Success {
			code = 1
		}
		return output, code
	case "text":
		output := formatText(result, noColor)
		code := 0
		if !result.Success {
			code = 1
		}
		return output, code
	case "quiet":
		if result.Success {
			return "", 0
		}
		return "", 1
	default:
		return fmt.Sprintf("Error: unknown format %q\n", format), 2
	}
}

func outputJSON(result anyvali.ParseResult) int {
	fmt.Print(formatJSON(result))
	if result.Success {
		return 0
	}
	return 1
}

type jsonOutput struct {
	Valid  bool        `json:"valid"`
	Data   any         `json:"data,omitempty"`
	Issues []jsonIssue `json:"issues,omitempty"`
}

type jsonIssue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
	Expected string `json:"expected,omitempty"`
	Received string `json:"received,omitempty"`
}

func formatJSON(result anyvali.ParseResult) string {
	out := jsonOutput{
		Valid: result.Success,
	}
	if result.Success {
		out.Data = result.Data
	} else {
		issues := make([]jsonIssue, len(result.Issues))
		for i, issue := range result.Issues {
			issues[i] = jsonIssue{
				Code:     issue.Code,
				Message:  issue.Message,
				Path:     formatPath(issue.Path),
				Expected: issue.Expected,
				Received: issue.Received,
			}
		}
		out.Issues = issues
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	return string(data) + "\n"
}

func outputText(result anyvali.ParseResult, noColor bool) int {
	fmt.Print(formatText(result, noColor))
	if result.Success {
		return 0
	}
	return 1
}

func formatText(result anyvali.ParseResult, noColor bool) string {
	var sb strings.Builder
	if result.Success {
		if noColor {
			sb.WriteString("Valid\n")
		} else {
			sb.WriteString("\u2713 Valid\n")
		}
	} else {
		count := len(result.Issues)
		if noColor {
			sb.WriteString(fmt.Sprintf("Invalid (%d issues)\n", count))
		} else {
			sb.WriteString(fmt.Sprintf("\u2717 Invalid (%d issues)\n", count))
		}
		for _, issue := range result.Issues {
			path := formatPath(issue.Path)
			if path == "" {
				path = "(root)"
			}
			sb.WriteString(fmt.Sprintf("  %s: [%s] %s\n", path, issue.Code, issue.Message))
		}
	}
	return sb.String()
}

func formatPath(path []any) string {
	if len(path) == 0 {
		return ""
	}
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = fmt.Sprintf("%v", p)
	}
	return strings.Join(parts, ".")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
