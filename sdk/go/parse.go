package anyvali

import (
	"encoding/json"
	"fmt"
	"math"
)

// baseSchema provides common schema functionality: coercions, defaults.
type baseSchema struct {
	coercions    []CoercionType
	defaultValue any
	hasDefault   bool
}

// SetDefault sets a default value for the schema.
func (b *baseSchema) setDefault(value any) {
	b.defaultValue = value
	b.hasDefault = true
}

// SetCoerce adds a coercion to the schema.
func (b *baseSchema) addCoercion(c CoercionType) {
	b.coercions = append(b.coercions, c)
}

// runPipeline executes the 5-step parse pipeline.
// validateFn performs the actual schema-specific validation and returns (parsedValue, issues).
func (b *baseSchema) runPipeline(input any, validateFn func(any) (any, []ValidationIssue)) ParseResult {
	value := input

	// Step 1: presence detection
	if value == nil || isAbsent(value) {
		// Step 3: apply default if absent
		if b.hasDefault {
			value = b.defaultValue
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
