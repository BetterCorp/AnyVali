package anyvali

import (
	"encoding/json"
	"fmt"
	"math"
)

// baseSchema provides common schema functionality: coercions, defaults, metadata.
type baseSchema struct {
	coercions    []CoercionType
	defaultValue any
	hasDefault   bool
	metadata     map[string]any
}

var reservedMetadataKeys = map[string]bool{
	"title": true, "description": true, "deprecated": true,
	"deprecatedMessage": true, "notStable": true, "since": true,
	"sensitive": true, "readonly": true, "writeonly": true, "examples": true,
}

// DescribeOpts contains optional reserved metadata fields for Describe().
type DescribeOpts struct {
	Title             string
	Deprecated        bool
	DeprecatedMessage string
	NotStable         bool
	Since             string
	Sensitive         bool
	Readonly          bool
	Writeonly         bool
	Examples          []any
}

// MetadataOpts contains options for Metadata().
type MetadataOpts struct {
	Replace bool
}

// deepCopyDefault returns a deep copy of a portable default value so that
// mutable containers are not shared across parses. Portable defaults are
// JSON-representable, so only maps and slices need recursive copying;
// everything else is immutable or copied by value.
func deepCopyDefault(v any) any {
	switch val := v.(type) {
	case map[string]any:
		cp := make(map[string]any, len(val))
		for k, item := range val {
			cp[k] = deepCopyDefault(item)
		}
		return cp
	case []any:
		cp := make([]any, len(val))
		for i, item := range val {
			cp[i] = deepCopyDefault(item)
		}
		return cp
	default:
		return val
	}
}

// SetDefault sets a default value for the schema.
func (b *baseSchema) setDefault(value any) {
	b.defaultValue = value
	b.hasDefault = true
}

func (b *baseSchema) defaultInfo() (any, bool) {
	return b.defaultValue, b.hasDefault
}

// SetCoerce adds a coercion to the schema.
func (b *baseSchema) addCoercion(c CoercionType) {
	b.coercions = append(b.coercions, c)
}

// runPipeline executes the 5-step parse pipeline.
// validateFn performs the actual schema-specific validation and returns (parsedValue, issues).
func (b *baseSchema) runPipeline(input any, validateFn func(any) (any, []ValidationIssue)) ParseResult {
	value := input
	usedDefault := false

	// Step 1: presence detection
	if value == nil || isAbsent(value) {
		// Step 3: apply default if absent
		if b.hasDefault {
			// Deep-copy so mutable defaults (maps/slices) are isolated per
			// parse. Pass-through schemas (Any/Unknown items, allowed unknown
			// keys) alias the stored default, so without copying a mutation to
			// one result would corrupt the default for the next parse.
			value = deepCopyDefault(b.defaultValue)
			usedDefault = true
		} else if value == nil {
			// nil is a present value (null), pass through to validation
			value = nil
		} else {
			// truly absent
			value = nil
		}
	} else {
		// Step 2: coercion (only for present values)
		for _, c := range b.coercions {
			coerced, err := applyCoercion(value, c)
			if err != nil {
				return ParseResult{
					Success: false,
					Issues: []ValidationIssue{
						{
							Code:     IssueCoercionFailed,
							Message:  fmt.Sprintf("coercion %s failed: %s", c, err.Error()),
							Received: fmt.Sprintf("%v", value),
						},
					},
				}
			}
			value = coerced
		}
	}

	// Step 4: validate
	parsed, issues := validateFn(value)
	if usedDefault && len(issues) > 0 {
		first := issues[0]
		issues = []ValidationIssue{{
			Code:     IssueDefaultInvalid,
			Message:  "default value is invalid",
			Path:     first.Path,
			Expected: first.Expected,
			Received: first.Received,
		}}
	}

	// Step 5: return result
	if len(issues) > 0 {
		return ParseResult{
			Success: false,
			Issues:  issues,
		}
	}
	return ParseResult{
		Success: true,
		Data:    parsed,
	}
}

func (b *baseSchema) addCoercionNode(node map[string]any) {
	if len(b.coercions) > 0 {
		cs := make([]any, len(b.coercions))
		for i, c := range b.coercions {
			cs[i] = string(c)
		}
		node["coerce"] = cs
	}
}

func (b *baseSchema) addDefaultNode(node map[string]any) {
	if b.hasDefault {
		node["default"] = b.defaultValue
	}
}

func (b *baseSchema) setDescribe(description string, opts *DescribeOpts) {
	if b.metadata == nil {
		b.metadata = make(map[string]any)
	}
	b.metadata["description"] = description
	if opts != nil {
		if opts.Title != "" {
			b.metadata["title"] = opts.Title
		}
		if opts.Deprecated {
			b.metadata["deprecated"] = true
		}
		if opts.DeprecatedMessage != "" {
			if !opts.Deprecated {
				panic("describe(): deprecatedMessage requires deprecated to be true")
			}
			b.metadata["deprecatedMessage"] = opts.DeprecatedMessage
		}
		if opts.NotStable {
			b.metadata["notStable"] = true
		}
		if opts.Since != "" {
			b.metadata["since"] = opts.Since
		}
		if opts.Sensitive {
			b.metadata["sensitive"] = true
		}
		if opts.Readonly {
			b.metadata["readonly"] = true
		}
		if opts.Writeonly {
			b.metadata["writeonly"] = true
		}
		if opts.Readonly && opts.Writeonly {
			panic("describe(): readonly and writeonly cannot both be true")
		}
		if opts.Examples != nil {
			b.metadata["examples"] = opts.Examples
		}
	}
}

func (b *baseSchema) setMetadata(meta map[string]any, opts *MetadataOpts) {
	for key := range meta {
		if reservedMetadataKeys[key] {
			panic(fmt.Sprintf("metadata(): %q is a reserved key. Use Describe() instead.", key))
		}
	}
	if opts != nil && opts.Replace {
		// Keep reserved keys, replace rest
		preserved := make(map[string]any)
		for k, v := range b.metadata {
			if reservedMetadataKeys[k] {
				preserved[k] = v
			}
		}
		b.metadata = preserved
		for k, v := range meta {
			b.metadata[k] = v
		}
	} else {
		if b.metadata == nil {
			b.metadata = make(map[string]any)
		}
		for k, v := range meta {
			b.metadata[k] = v
		}
	}
}

func (b *baseSchema) addMetadataNode(node map[string]any) {
	if len(b.metadata) > 0 {
		metaCopy := make(map[string]any, len(b.metadata))
		for k, v := range b.metadata {
			metaCopy[k] = v
		}
		node["metadata"] = metaCopy
	}
}

// Helper: extract a float64 from any numeric input.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

// toInt64 extracts an int64 from any numeric input, checking for integrality.
func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		if n > math.MaxInt64 {
			return 0, false
		}
		return int64(n), true
	case float32:
		if n != float32(int64(n)) {
			return 0, false
		}
		return int64(n), true
	case float64:
		if n != float64(int64(n)) {
			return 0, false
		}
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

// toUint64 extracts a uint64 from any numeric input, checking for non-negativity and integrality.
func toUint64(v any) (uint64, bool) {
	switch n := v.(type) {
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int8:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int16:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int32:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case uint:
		return uint64(n), true
	case uint8:
		return uint64(n), true
	case uint16:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case float32:
		if n < 0 || n != float32(uint64(n)) {
			return 0, false
		}
		return uint64(n), true
	case float64:
		if n < 0 || n != float64(uint64(n)) {
			return 0, false
		}
		return uint64(n), true
	case json.Number:
		// Try int64 first
		i, err := n.Int64()
		if err == nil && i >= 0 {
			return uint64(i), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func typeName(v any) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case bool:
		return "bool"
	case string:
		return "string"
	case float64, float32:
		return "number"
	case int, int8, int16, int32, int64:
		return "int"
	case uint, uint8, uint16, uint32, uint64:
		return "uint"
	case json.Number:
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}
