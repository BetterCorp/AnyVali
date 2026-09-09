package anyvali

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
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
			if previous, ok := defs[name]; ok {
				equal, err := equivalentDefinition(previous, node)
				if err != nil {
					return err
				}
				if !equal {
					return fmt.Errorf("conflicting definition %q", name)
				}
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

// Compare JSON numbers exactly, independently of native Go numeric types.
func equivalentDefinition(left, right map[string]any) (bool, error) {
	normalize := func(value any) (any, error) {
		data, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		var normalized any
		err = decoder.Decode(&normalized)
		return normalized, err
	}
	a, err := normalize(left)
	if err != nil {
		return false, err
	}
	b, err := normalize(right)
	if err != nil {
		return false, err
	}
	return jsonValuesEqual(a, b), nil
}

func jsonValuesEqual(left, right any) bool {
	switch a := left.(type) {
	case json.Number:
		b, ok := right.(json.Number)
		if !ok {
			return false
		}
		ar, aok := new(big.Rat).SetString(string(a))
		br, bok := new(big.Rat).SetString(string(b))
		return aok && bok && ar.Cmp(br) == 0
	case map[string]any:
		b, ok := right.(map[string]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for key, value := range a {
			other, present := b[key]
			if !present || !jsonValuesEqual(value, other) {
				return false
			}
		}
		return true
	case []any:
		b, ok := right.([]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for i, value := range a {
			if !jsonValuesEqual(value, b[i]) {
				return false
			}
		}
		return true
	default:
		return left == right
	}
}
