package anyvali

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// importContext owns reference resolution for one document. Placeholders are
// registered before descending so recursive references share the completed graph.
type importContext struct {
	refs map[string]*RefSchema
}

// Import converts a Document to a Schema.
func Import(doc *Document) (Schema, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}
	ctx := &importContext{refs: make(map[string]*RefSchema)}
	return ctx.importNode(doc.Root, doc.Definitions)
}

// ImportJSON parses JSON bytes into a Document and then converts to a Schema.
func ImportJSON(data []byte) (Schema, error) {
	var doc Document
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("failed to parse JSON: trailing value or invalid content")
	}
	return Import(&doc)
}

func (ctx *importContext) importNode(node map[string]any, defs map[string]map[string]any) (Schema, error) {
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
		return ctx.importArraySchema(node, defs)
	case "tuple":
		return ctx.importTupleSchema(node, defs)
	case "object":
		return ctx.importObjectSchema(node, defs)
	case "record":
		return ctx.importRecordSchema(node, defs)
	case "union":
		return ctx.importUnionSchema(node, defs)
	case "intersection":
		return ctx.importIntersectionSchema(node, defs)
	case "optional":
		return ctx.importOptionalSchema(node, defs)
	case "nullable":
		return ctx.importNullableSchema(node, defs)
	case "ref":
		return ctx.importRefSchema(node, defs)
	default:
		return nil, fmt.Errorf("unsupported schema kind: %s", kind)
	}
}

func importStringSchema(node map[string]any) (*StringSchema, error) {
	s := String()
	if v, ok, err := getCountConstraint(node, "minLength"); err != nil {
		return nil, err
	} else if ok {
		s.MinLength(v)
	}
	if v, ok, err := getCountConstraint(node, "maxLength"); err != nil {
		return nil, err
	} else if ok {
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
	if err := importNumericConstraintsFloat(s, node); err != nil {
		return nil, err
	}
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importFloat32Schema(node map[string]any) (*Float32Schema, error) {
	s := Float32()
	if err := importNumericConstraintsFloat(s, node); err != nil {
		return nil, err
	}
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importNumericConstraintsFloat(s *Float64Schema, node map[string]any) error {
	for _, key := range []string{"min", "max", "exclusiveMin", "exclusiveMax", "multipleOf"} {
		value, present := node[key]
		if !present {
			continue
		}
		n, ok := toFloat64(value)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
			return fmt.Errorf("%s must be a finite float64 number", key)
		}
		switch key {
		case "min":
			s.Min(n)
		case "max":
			s.Max(n)
		case "exclusiveMin":
			s.ExclusiveMin(n)
		case "exclusiveMax":
			s.ExclusiveMax(n)
		case "multipleOf":
			s.MultipleOf(n)
		}
	}
	return nil
}

func importIntSchema(node map[string]any, kind string, low, high int64) (*IntSchema, error) {
	s := newIntSchema(kind, low, high)
	if err := importNumericConstraintsInt(s, node); err != nil {
		return nil, err
	}
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importUintSchema(node map[string]any, kind string, low, high uint64) (*IntSchema, error) {
	s := newUintSchema(kind, low, high)
	if err := importNumericConstraintsInt(s, node); err != nil {
		return nil, err
	}
	importCoercions(&s.baseSchema, node)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func importNumericConstraintsInt(s *IntSchema, node map[string]any) error {
	for _, key := range []string{"min", "max", "exclusiveMin", "exclusiveMax", "multipleOf"} {
		value, present := node[key]
		if !present {
			continue
		}
		rat, ok := exactNumberRat(value)
		if !ok {
			return fmt.Errorf("%s must be a finite numeric constraint", key)
		}
		if s.exactConstraints == nil {
			s.exactConstraints = make(map[string]exactConstraint)
		}
		s.exactConstraints[key] = exactConstraint{value: value, rat: rat}
	}
	return nil
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

func (ctx *importContext) importArraySchema(node map[string]any, defs map[string]map[string]any) (*ArraySchema, error) {
	// Accept both "item" and "items" keys for compatibility
	itemNode, ok := node["item"].(map[string]any)
	if !ok {
		itemNode, ok = node["items"].(map[string]any)
	}
	if !ok {
		return nil, fmt.Errorf("array schema missing 'item' or 'items' field")
	}
	item, err := ctx.importNode(itemNode, defs)
	if err != nil {
		return nil, fmt.Errorf("array item: %w", err)
	}
	s := Array(item)
	if v, ok, err := getCountConstraint(node, "minItems"); err != nil {
		return nil, err
	} else if ok {
		s.MinItems(v)
	}
	if v, ok, err := getCountConstraint(node, "maxItems"); err != nil {
		return nil, err
	} else if ok {
		s.MaxItems(v)
	}
	importDefault(&s.baseSchema, node)
	return s, nil
}

func (ctx *importContext) importTupleSchema(node map[string]any, defs map[string]map[string]any) (*TupleSchema, error) {
	raw, present := node["elements"]
	if !present {
		raw = node["items"]
	}
	itemNodes, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("tuple schema missing or invalid 'elements' field")
	}
	items := make([]Schema, len(itemNodes))
	for i, itemNode := range itemNodes {
		m, ok := itemNode.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tuple item %d is not a valid schema node", i)
		}
		item, err := ctx.importNode(m, defs)
		if err != nil {
			return nil, fmt.Errorf("tuple item %d: %w", i, err)
		}
		items[i] = item
	}
	return Tuple(items...), nil
}

func (ctx *importContext) importObjectSchema(node map[string]any, defs map[string]map[string]any) (*ObjectSchema, error) {
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
		schema, err := ctx.importNode(m, defs)
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

func (ctx *importContext) importRecordSchema(node map[string]any, defs map[string]map[string]any) (*RecordSchema, error) {
	raw, present := node["values"]
	if !present {
		raw, present = node["valueSchema"]
		if !present {
			raw = node["value"]
		}
	} // Compatibility aliases; exports follow the specification.
	valueNode, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("record schema missing or invalid 'values' field")
	}
	value, err := ctx.importNode(valueNode, defs)
	if err != nil {
		return nil, fmt.Errorf("record value: %w", err)
	}
	return Record(value), nil
}

func (ctx *importContext) importUnionSchema(node map[string]any, defs map[string]map[string]any) (*UnionSchema, error) {
	// Accept both "schemas" and "variants" keys for compatibility
	schemasNode, ok := node["schemas"].([]any)
	if !ok {
		schemasNode, ok = node["variants"].([]any)
	}
	if !ok {
		return nil, fmt.Errorf("union schema missing 'schemas' or 'variants' field")
	}
	schemas := make([]Schema, len(schemasNode))
	for i, sn := range schemasNode {
		m, ok := sn.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("union schema %d is not a valid schema node", i)
		}
		schema, err := ctx.importNode(m, defs)
		if err != nil {
			return nil, fmt.Errorf("union schema %d: %w", i, err)
		}
		schemas[i] = schema
	}
	return Union(schemas...), nil
}

func (ctx *importContext) importIntersectionSchema(node map[string]any, defs map[string]map[string]any) (*IntersectionSchema, error) {
	raw, present := node["allOf"]
	if !present {
		raw = node["schemas"]
	}
	schemasNode, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("intersection schema missing or invalid 'allOf' field")
	}
	schemas := make([]Schema, len(schemasNode))
	for i, sn := range schemasNode {
		m, ok := sn.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("intersection schema %d is not a valid schema node", i)
		}
		schema, err := ctx.importNode(m, defs)
		if err != nil {
			return nil, fmt.Errorf("intersection schema %d: %w", i, err)
		}
		schemas[i] = schema
	}
	return Intersection(schemas...), nil
}

func (ctx *importContext) importOptionalSchema(node map[string]any, defs map[string]map[string]any) (*OptionalSchema, error) {
	schemaNode, ok := node["schema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("optional schema missing 'schema' field")
	}
	inner, err := ctx.importNode(schemaNode, defs)
	if err != nil {
		return nil, fmt.Errorf("optional inner: %w", err)
	}
	s := Optional(inner)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func (ctx *importContext) importNullableSchema(node map[string]any, defs map[string]map[string]any) (*NullableSchema, error) {
	schemaNode, ok := node["schema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("nullable schema missing 'schema' field")
	}
	inner, err := ctx.importNode(schemaNode, defs)
	if err != nil {
		return nil, fmt.Errorf("nullable inner: %w", err)
	}
	s := Nullable(inner)
	importDefault(&s.baseSchema, node)
	return s, nil
}

func (ctx *importContext) importRefSchema(node map[string]any, defs map[string]map[string]any) (*RefSchema, error) {
	ref, ok := node["ref"].(string)
	if !ok {
		return nil, fmt.Errorf("ref schema missing 'ref' field")
	}
	if cached, ok := ctx.refs[ref]; ok {
		return cached, nil
	}
	s := newRefSchema(ref)
	ctx.refs[ref] = s
	const defPrefix = "#/definitions/"
	if len(ref) > len(defPrefix) && ref[:len(defPrefix)] == defPrefix {
		if defNode, ok := defs[ref[len(defPrefix):]]; ok {
			resolved, err := ctx.importNode(defNode, defs)
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

func getCountConstraint(node map[string]any, key string) (int, bool, error) {
	v, ok := node[key]
	if !ok {
		return 0, false, nil
	}
	rat, ok := exactNumberRat(v)
	if !ok || !rat.IsInt() || !rat.Num().IsInt64() {
		return 0, false, fmt.Errorf("%s must be a nonnegative integer within the platform int range", key)
	}
	n := rat.Num().Int64()
	if n < 0 || int64(int(n)) != n {
		return 0, false, fmt.Errorf("%s must be a nonnegative integer within the platform int range", key)
	}
	return int(n), true, nil
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
