package anyvali

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Export converts a Schema to a Document.
func Export(schema Schema, mode ExportMode) (*Document, error) {
	node := schema.ToNode()

	doc := &Document{
		AnyvaliVersion: "1.0",
		SchemaVersion:  "1.1",
		Root:           node,
		Definitions:    make(map[string]map[string]any),
		Extensions:     make(map[string]any),
	}

	// Collect any ref definitions from the schema tree
	if err := collectSchemaDefinitions(schema, doc.Definitions, make(map[*RefSchema]bool)); err != nil {
		return nil, err
	}

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

// Walk the schema graph, stopping at previously visited nodes for recursion.
func collectSchemaDefinitions(schema Schema, defs map[string]map[string]any, seen map[*RefSchema]bool) error {
	var children []Schema
	switch s := schema.(type) {
	case *RefSchema:
		if seen[s] {
			return nil
		}
		seen[s] = true
		if s.resolved == nil {
			return nil
		}
		if strings.HasPrefix(s.ref, "#/definitions/") {
			name := strings.TrimPrefix(s.ref, "#/definitions/")
			node := s.resolved.ToNode()
			if previous, ok := defs[name]; ok && !reflect.DeepEqual(previous, node) {
				return fmt.Errorf("conflicting definition %q", name)
			}
			defs[name] = node
		}
		children = []Schema{s.resolved}
	case *ObjectSchema:
		for _, child := range s.properties {
			children = append(children, child)
		}
	case *ArraySchema:
		children = []Schema{s.item}
	case *RecordSchema:
		children = []Schema{s.valueSchema}
	case *TupleSchema:
		children = s.items
	case *UnionSchema:
		children = s.schemas
	case *IntersectionSchema:
		children = s.schemas
	case *OptionalSchema:
		children = []Schema{s.inner}
	case *NullableSchema:
		children = []Schema{s.inner}
	}
	for _, child := range children {
		if err := collectSchemaDefinitions(child, defs, seen); err != nil {
			return err
		}
	}
	return nil
}
