package anyvali

import (
	"encoding/json"
	"fmt"
	"math"
)

// Import converts a Document to a Schema.
func Import(doc *Document) (Schema, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}
	return importNode(doc.Root, doc.Definitions)
}

// ImportJSON parses JSON bytes into a Document and then converts to a Schema.
func ImportJSON(data []byte) (Schema, error) {
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return Import(&doc)
}

func importNode(node map[string]any, defs map[string]map[string]any) (Schema, error) {
	if node == nil {
		return nil, fmt.Errorf("schema node is nil")
	}

	kind, ok := node["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("schema node missing 'kind' field")
	}

	switch kind {
	case "string":
		return importStringSchema(node)
	case "number", "float64":
		return importFloat64Schema(node, kind)
	case "float32":
		return importFloat32Schema(node)
	case "int", "int64":
		return importIntSchema(node, kind, math.MinInt64, math.MaxInt64)
	case "int8":
		return importIntSchema(node, kind, math.MinInt8, math.MaxInt8)
	case "int16":
		return importIntSchema(node, kind, math.MinInt16, math.MaxInt16)
	case "int32":
		return importIntSchema(node, kind, math.MinInt32, math.MaxInt32)
	case "uint8":
		return importUintSchema(node, kind, 0, math.MaxUint8)
	case "uint16":
		return importUintSchema(node, kind, 0, math.MaxUint16)
	case "uint32":
		return importUintSchema(node, kind, 0, math.MaxUint32)
	case "uint64":
		return importUintSchema(node, kind, 0, math.MaxUint64)
	case "bool":
		return importBoolSchema(node)
	case "null":
		return Null(), nil
	case "any":
		return Any(), nil
	case "unknown":
		return Unknown(), nil
	case "never":
		return Never(), nil
	case "literal":
		return importLiteralSchema(node)
	case "enum":
		return importEnumSchema(node)
	case "array":
		return importArraySchema(node, defs)
	case "tuple":
		return importTupleSchema(node, defs)
	case "object":
		return importObjectSchema(node, defs)
	case "record":
		return importRecordSchema(node, defs)
	case "union":
		return importUnionSchema(node, defs)
	case "intersection":
		return importIntersectionSchema(node, defs)
	case "optional":
		return importOptionalSchema(node, defs)
	case "nullable":
		return importNullableSchema(node, defs)
	case "ref":
		return importRefSchema(node, defs)
	default:
		return nil, fmt.Errorf("unsupported schema kind: %s", kind)
	}
}

func importStringSchema(node map[string]any) (*StringSchema, error) {
	s := String()
	if v, ok := getInt(node, "minLength"); ok {
		s.MinLength(v)
	}
	if v, ok := getInt(node, "maxLength"); ok {
		s.MaxLength(v)
	}
	if v, ok := node["pattern"].(string); ok {
		s.Pattern(v)
	}
	if v, ok := node["startsWith"].(string); ok {
		s.StartsWith(v)
	}
	if v, ok := node["endsWith"].(string); ok {
		s.EndsWith(v)
	}
	if v, ok := node["includes"].(string); ok {
		s.Includes(v)
	}
	if v, ok := node["format"].(string); ok {
		s.Format(v)
	}
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importFloat64Schema(node map[string]any, kind string) (*Float64Schema, error) {
	s := newFloat64Schema(kind)
	importNumericConstraintsFloat(s, node)
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importFloat32Schema(node map[string]any) (*Float32Schema, error) {
	s := Float32()
	importNumericConstraintsFloat(s, node)
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importNumericConstraintsFloat(s *Float64Schema, node map[string]any) {
	if v, ok := getFloat(node, "min"); ok {
		s.Min(v)
	}
	if v, ok := getFloat(node, "max"); ok {
		s.Max(v)
	}
	if v, ok := getFloat(node, "exclusiveMin"); ok {
		s.ExclusiveMin(v)
	}
	if v, ok := getFloat(node, "exclusiveMax"); ok {
		s.ExclusiveMax(v)
	}
	if v, ok := getFloat(node, "multipleOf"); ok {
		s.MultipleOf(v)
	}
}

func importIntSchema(node map[string]any, kind string, low, high int64) (*IntSchema, error) {
	s := newIntSchema(kind, low, high)
	importNumericConstraintsInt(s, node)
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importUintSchema(node map[string]any, kind string, low, high uint64) (*IntSchema, error) {
	s := newUintSchema(kind, low, high)
	importNumericConstraintsInt(s, node)
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importNumericConstraintsInt(s *IntSchema, node map[string]any) {
	if v, ok := getFloat(node, "min"); ok {
		iv := int64(v)
		s.Min(iv)
	}
	if v, ok := getFloat(node, "max"); ok {
		iv := int64(v)
		s.Max(iv)
	}
	if v, ok := getFloat(node, "exclusiveMin"); ok {
		iv := int64(v)
		s.ExclusiveMin(iv)
	}
	if v, ok := getFloat(node, "exclusiveMax"); ok {
		iv := int64(v)
		s.ExclusiveMax(iv)
	}
	if v, ok := getFloat(node, "multipleOf"); ok {
		iv := int64(v)
		s.MultipleOf(iv)
	}
}

func importBoolSchema(node map[string]any) (*BoolSchema, error) {
	s := Bool()
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importLiteralSchema(node map[string]any) (*LiteralSchema, error) {
	value, ok := node["value"]
	if !ok {
		return nil, fmt.Errorf("literal schema missing 'value' field")
	}
	return Literal(value), nil
}

func importEnumSchema(node map[string]any) (*EnumSchema, error) {
	values, ok := node["values"].([]any)
	if !ok {
		return nil, fmt.Errorf("enum schema missing 'values' field")
	}
	s := Enum(values...)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importArraySchema(node map[string]any, defs map[string]map[string]any) (*ArraySchema, error) {
	itemNode, ok := node["item"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("array schema missing 'item' field")
	}
	item, err := importNode(itemNode, defs)
	if err != nil {
		return nil, fmt.Errorf("array item: %w", err)
	}
	s := Array(item)
	if v, ok := getInt(node, "minItems"); ok {
		s.MinItems(v)
	}
	if v, ok := getInt(node, "maxItems"); ok {
		s.MaxItems(v)
	}
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importTupleSchema(node map[string]any, defs map[string]map[string]any) (*TupleSchema, error) {
	itemNodes, ok := node["items"].([]any)
	if !ok {
		return nil, fmt.Errorf("tuple schema missing 'items' field")
	}
	items := make([]Schema, len(itemNodes))
	for i, itemNode := range itemNodes {
		m, ok := itemNode.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tuple item %d is not a valid schema node", i)
		}
		item, err := importNode(m, defs)
		if err != nil {
			return nil, fmt.Errorf("tuple item %d: %w", i, err)
		}
		items[i] = item
	}
	return Tuple(items...), nil
}

func importObjectSchema(node map[string]any, defs map[string]map[string]any) (*ObjectSchema, error) {
	propsNode, ok := node["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("object schema missing 'properties' field")
	}
	props := make(map[string]Schema)
	for key, val := range propsNode {
		m, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("property %q is not a valid schema node", key)
		}
		schema, err := importNode(m, defs)
		if err != nil {
			return nil, fmt.Errorf("property %q: %w", key, err)
		}
		props[key] = schema
	}
	s := Object(props)

	// Set required fields
	if reqList, ok := node["required"].([]any); ok {
		reqs := make([]string, 0, len(reqList))
		for _, r := range reqList {
			if rs, ok := r.(string); ok {
				reqs = append(reqs, rs)
			}
		}
		s.Required(reqs...)
	}

	// Set unknown keys mode
	if mode, ok := node["unknownKeys"].(string); ok {
		s.UnknownKeys(UnknownKeyMode(mode))
	}

	importDefault(&s.baseSchema, node)
	return s, nil
}

func importRecordSchema(node map[string]any, defs map[string]map[string]any) (*RecordSchema, error) {
	valueNode, ok := node["value"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("record schema missing 'value' field")
	}
	value, err := importNode(valueNode, defs)
	if err != nil {
		return nil, fmt.Errorf("record value: %w", err)
	}
	return Record(value), nil
}

func importUnionSchema(node map[string]any, defs map[string]map[string]any) (*UnionSchema, error) {
	schemasNode, ok := node["schemas"].([]any)
	if !ok {
		return nil, fmt.Errorf("union schema missing 'schemas' field")
	}
	schemas := make([]Schema, len(schemasNode))
	for i, sn := range schemasNode {
		m, ok := sn.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("union schema %d is not a valid schema node", i)
		}
		schema, err := importNode(m, defs)
		if err != nil {
			return nil, fmt.Errorf("union schema %d: %w", i, err)
		}
		schemas[i] = schema
	}
	return Union(schemas...), nil
}

func importIntersectionSchema(node map[string]any, defs map[string]map[string]any) (*IntersectionSchema, error) {
	schemasNode, ok := node["schemas"].([]any)
	if !ok {
		return nil, fmt.Errorf("intersection schema missing 'schemas' field")
	}
	schemas := make([]Schema, len(schemasNode))
	for i, sn := range schemasNode {
		m, ok := sn.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("intersection schema %d is not a valid schema node", i)
		}
		schema, err := importNode(m, defs)
		if err != nil {
			return nil, fmt.Errorf("intersection schema %d: %w", i, err)
		}
		schemas[i] = schema
	}
	return Intersection(schemas...), nil
}

func importOptionalSchema(node map[string]any, defs map[string]map[string]any) (*OptionalSchema, error) {
	schemaNode, ok := node["schema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("optional schema missing 'schema' field")
	}
	inner, err := importNode(schemaNode, defs)
	if err != nil {
		return nil, fmt.Errorf("optional inner: %w", err)
	}
	s := Optional(inner)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importNullableSchema(node map[string]any, defs map[string]map[string]any) (*NullableSchema, error) {
	schemaNode, ok := node["schema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("nullable schema missing 'schema' field")
	}
	inner, err := importNode(schemaNode, defs)
	if err != nil {
		return nil, fmt.Errorf("nullable inner: %w", err)
	}
	s := Nullable(inner)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importRefSchema(node map[string]any, defs map[string]map[string]any) (*RefSchema, error) {
	ref, ok := node["ref"].(string)
	if !ok {
		return nil, fmt.Errorf("ref schema missing 'ref' field")
	}
	s := newRefSchema(ref)

	// Try to resolve from definitions
	// ref format: #/definitions/Name
	const defPrefix = "#/definitions/"
	if len(ref) > len(defPrefix) && ref[:len(defPrefix)] == defPrefix {
		defName := ref[len(defPrefix):]
		if defNode, ok := defs[defName]; ok {
			resolved, err := importNode(defNode, defs)
			if err != nil {
				return nil, fmt.Errorf("ref %q: %w", ref, err)
			}
			s.Resolve(resolved)
		}
	}

	return s, nil
}

// Helper functions for extracting typed values from node maps.

func getFloat(node map[string]any, key string) (float64, bool) {
	v, ok := node[key]
	if !ok {
		return 0, false
	}
	f, ok := toFloat64(v)
	return f, ok
}

func getInt(node map[string]any, key string) (int, bool) {
	v, ok := node[key]
	if !ok {
		return 0, false
	}
	f, fOk := toFloat64(v)
	if fOk {
		return int(f), true
	}
	return 0, false
}

func importCoercions(b *baseSchema, node map[string]any) {
	cs, ok := node["coerce"].([]any)
	if !ok {
		return
	}
	for _, c := range cs {
		if s, ok := c.(string); ok {
			b.addCoercion(CoercionType(s))
		}
	}
}

func importDefault(b *baseSchema, node map[string]any) {
	v, ok := node["default"]
	if !ok {
		return
	}
	b.setDefault(v)
}
