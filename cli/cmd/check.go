package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const checkUsage = `Usage: anyvali check <schema-file>

Validates that an AnyVali schema document is valid and portable.

Flags:
  -h, --help    Show help
`

// RunCheck executes the check command and returns the exit code.
func RunCheck(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(checkUsage)
		if len(args) == 0 {
			return 2
		}
		return 0
	}

	schemaFile := args[0]
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read schema file: %s\n", err)
		return 2
	}

	output, code := CheckSchema(data)
	fmt.Print(output)
	return code
}

// CheckSchema validates schema bytes and returns the output string and exit code.
func CheckSchema(data []byte) (string, int) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Sprintf("Error: invalid JSON: %s\n", err), 2
	}

	root, ok := doc["root"].(map[string]any)
	if !ok {
		return "Error: schema document missing 'root' field\n", 2
	}

	kind, _ := root["kind"].(string)
	if kind == "" {
		return "Error: root schema missing 'kind' field\n", 2
	}

	var issues []string

	// Check for non-portable extensions
	extensions, _ := doc["extensions"].(map[string]any)
	for extName, extVal := range extensions {
		if m, ok := extVal.(map[string]any); ok && len(m) > 0 {
			issues = append(issues, fmt.Sprintf("Extension %q contains semantic extensions", extName))
		}
	}

	// Walk the schema tree for portability issues
	walkPortability(root, "", &issues)

	// Count properties and definitions
	propCount := countProperties(root)
	defs, _ := doc["definitions"].(map[string]any)
	defCount := len(defs)
	extCount := len(extensions)

	var sb strings.Builder

	if len(issues) == 0 {
		sb.WriteString("\u2713 Schema is valid and portable\n")
		sb.WriteString(fmt.Sprintf("  Kind: %s\n", kind))
		sb.WriteString(fmt.Sprintf("  Properties: %d\n", propCount))
		sb.WriteString(fmt.Sprintf("  Definitions: %d\n", defCount))
		sb.WriteString(fmt.Sprintf("  Extensions: %d\n", extCount))
		return sb.String(), 0
	}

	sb.WriteString("\u2717 Schema has portability issues\n")
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("  - %s\n", issue))
	}
	return sb.String(), 1
}

func walkPortability(node map[string]any, path string, issues *[]string) {
	if node == nil {
		return
	}

	// Check for refine (custom validators)
	if _, ok := node["refine"]; ok {
		p := path
		if p == "" {
			p = "(root)"
		}
		*issues = append(*issues, fmt.Sprintf("Custom validator at path %q is not portable", p))
	}

	// Check for transform
	if _, ok := node["transform"]; ok {
		p := path
		if p == "" {
			p = "(root)"
		}
		*issues = append(*issues, fmt.Sprintf("Transform at path %q is not portable", p))
	}

	// Recurse into properties
	if props, ok := node["properties"].(map[string]any); ok {
		for key, val := range props {
			if m, ok := val.(map[string]any); ok {
				childPath := key
				if path != "" {
					childPath = path + "." + key
				}
				walkPortability(m, childPath, issues)
			}
		}
	}

	// Recurse into item (array)
	if item, ok := node["item"].(map[string]any); ok {
		walkPortability(item, path+".[]", issues)
	}

	// Recurse into items (tuple)
	if items, ok := node["items"].([]any); ok {
		for i, item := range items {
			if m, ok := item.(map[string]any); ok {
				walkPortability(m, fmt.Sprintf("%s.[%d]", path, i), issues)
			}
		}
	}

	// Recurse into schemas (union, intersection)
	if schemas, ok := node["schemas"].([]any); ok {
		for i, s := range schemas {
			if m, ok := s.(map[string]any); ok {
				walkPortability(m, fmt.Sprintf("%s.<variant %d>", path, i), issues)
			}
		}
	}

	// Recurse into inner schema (optional, nullable)
	if inner, ok := node["schema"].(map[string]any); ok {
		walkPortability(inner, path, issues)
	}

	// Recurse into value (record)
	if val, ok := node["value"].(map[string]any); ok {
		walkPortability(val, path+".{}", issues)
	}
}

func countProperties(node map[string]any) int {
	if node == nil {
		return 0
	}
	props, ok := node["properties"].(map[string]any)
	if !ok {
		return 0
	}
	return len(props)
}
