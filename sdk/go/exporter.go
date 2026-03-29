package anyvali

import "encoding/json"

// Export converts a Schema to a Document.
func Export(schema Schema, mode ExportMode) (*Document, error) {
	node := schema.ToNode()

	doc := &Document{
		AnyvaliVersion: "1.0",
		SchemaVersion:  "1",
		Root:           node,
		Definitions:    make(map[string]map[string]any),
		Extensions:     make(map[string]any),
	}

	// Collect any ref definitions from the schema tree
	collectDefinitions(node, doc.Definitions)

	return doc, nil
}

// ExportJSON converts a Schema to JSON bytes.
func ExportJSON(schema Schema, mode ExportMode) ([]byte, error) {
	doc, err := Export(schema, mode)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(doc, "", "  ")
}

// collectDefinitions walks the schema node tree and collects any definitions.
func collectDefinitions(node map[string]any, defs map[string]map[string]any) {
	if node == nil {
		return
	}

	// Walk nested nodes
	for _, v := range node {
		switch val := v.(type) {
		case map[string]any:
			collectDefinitions(val, defs)
		case []any:
			for _, item := range val {
				if m, ok := item.(map[string]any); ok {
					collectDefinitions(m, defs)
				}
			}
		}
	}
}
