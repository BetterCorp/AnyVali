package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const inspectUsage = `Usage: anyvali inspect <schema-file>

Shows the structure of an AnyVali schema document.

Flags:
  -h, --help    Show help
`

// RunInspect executes the inspect command and returns the exit code.
func RunInspect(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(inspectUsage)
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

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid JSON: %s\n", err)
		return 2
	}

	root, ok := doc["root"].(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: schema document missing 'root' field\n")
		return 2
	}

	output := InspectNode(root, doc)
	fmt.Print(output)
	return 0
}

// InspectNode produces an inspection string for a schema node within a document.
func InspectNode(root map[string]any, doc map[string]any) string {
	var sb strings.Builder

	kind, _ := root["kind"].(string)
	sb.WriteString(fmt.Sprintf("Schema: %s\n", kind))

	switch kind {
	case "object":
		inspectObject(&sb, root)
	case "array":
		inspectArray(&sb, root)
	case "string":
		inspectString(&sb, root)
	case "union":
		inspectUnion(&sb, root)
	case "tuple":
		inspectTuple(&sb, root)
	default:
		inspectConstraints(&sb, root, "  ")
	}

	return sb.String()
}

func inspectObject(sb *strings.Builder, node map[string]any) {
	properties, _ := node["properties"].(map[string]any)
	requiredList, _ := node["required"].([]any)
	unknownKeys, _ := node["unknownKeys"].(string)

	requiredSet := make(map[string]bool)
	for _, r := range requiredList {
		if s, ok := r.(string); ok {
			requiredSet[s] = true
		}
	}

	if len(properties) > 0 {
		sb.WriteString("Properties:\n")

		// Sort property names for consistent output
		keys := make([]string, 0, len(properties))
		for k := range properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			prop, ok := properties[key].(map[string]any)
			if !ok {
				continue
			}
			propKind := resolveKind(prop)
			optionality := "optional"
			if requiredSet[key] {
				optionality = "required"
			}
			// Check if the property is wrapped in optional
			if prop["kind"] == "optional" {
				optionality = "optional"
				if inner, ok := prop["schema"].(map[string]any); ok {
					propKind = resolveKind(inner)
					prop = inner
				}
			}

			constraints := gatherConstraints(prop)
			constraintStr := ""
			if len(constraints) > 0 {
				constraintStr = " [" + strings.Join(constraints, ", ") + "]"
			}

			sb.WriteString(fmt.Sprintf("  %s: %s (%s)%s\n", key, propKind, optionality, constraintStr))
		}
	}

	if unknownKeys != "" {
		sb.WriteString(fmt.Sprintf("Unknown keys: %s\n", unknownKeys))
	}
}

func inspectArray(sb *strings.Builder, node map[string]any) {
	item, _ := node["item"].(map[string]any)
	itemKind := resolveKind(item)
	sb.WriteString(fmt.Sprintf("Items: %s\n", itemKind))
	constraints := gatherConstraints(node)
	for _, c := range constraints {
		sb.WriteString(fmt.Sprintf("  %s\n", c))
	}
}

func inspectString(sb *strings.Builder, node map[string]any) {
	constraints := gatherConstraints(node)
	if len(constraints) > 0 {
		sb.WriteString("Constraints:\n")
		for _, c := range constraints {
			sb.WriteString(fmt.Sprintf("  %s\n", c))
		}
	}
}

func inspectUnion(sb *strings.Builder, node map[string]any) {
	schemas, _ := node["schemas"].([]any)
	sb.WriteString("Variants:\n")
	for i, s := range schemas {
		if m, ok := s.(map[string]any); ok {
			sb.WriteString(fmt.Sprintf("  [%d]: %s\n", i, resolveKind(m)))
		}
	}
}

func inspectTuple(sb *strings.Builder, node map[string]any) {
	items, _ := node["items"].([]any)
	sb.WriteString("Items:\n")
	for i, item := range items {
		if m, ok := item.(map[string]any); ok {
			sb.WriteString(fmt.Sprintf("  [%d]: %s\n", i, resolveKind(m)))
		}
	}
}

func inspectConstraints(sb *strings.Builder, node map[string]any, prefix string) {
	constraints := gatherConstraints(node)
	if len(constraints) > 0 {
		sb.WriteString("Constraints:\n")
		for _, c := range constraints {
			sb.WriteString(fmt.Sprintf("%s%s\n", prefix, c))
		}
	}
}

func resolveKind(node map[string]any) string {
	if node == nil {
		return "unknown"
	}
	kind, _ := node["kind"].(string)

	switch kind {
	case "array":
		if item, ok := node["item"].(map[string]any); ok {
			return "array<" + resolveKind(item) + ">"
		}
		return "array"
	case "optional":
		if inner, ok := node["schema"].(map[string]any); ok {
			return resolveKind(inner)
		}
		return "optional"
	case "nullable":
		if inner, ok := node["schema"].(map[string]any); ok {
			return resolveKind(inner) + " | null"
		}
		return "nullable"
	case "record":
		if val, ok := node["value"].(map[string]any); ok {
			return "record<" + resolveKind(val) + ">"
		}
		return "record"
	default:
		return kind
	}
}

func gatherConstraints(node map[string]any) []string {
	var constraints []string

	constraintKeys := []string{
		"minLength", "maxLength", "pattern", "format",
		"startsWith", "endsWith", "includes",
		"min", "max", "exclusiveMin", "exclusiveMax", "multipleOf",
		"minItems", "maxItems",
	}

	for _, key := range constraintKeys {
		if v, ok := node[key]; ok {
			constraints = append(constraints, fmt.Sprintf("%s=%v", key, v))
		}
	}

	return constraints
}
