package anyvali

// sentinel value to indicate absence of input
type absent struct{}

var absentValue = absent{}

// isAbsent checks if a value represents an absent input.
func isAbsent(v any) bool {
	_, ok := v.(absent)
	return ok
}

// applyDefault returns the default value if the input is absent, otherwise returns the input.
func applyDefault(value any, defaultValue any, hasDefault bool) any {
	if isAbsent(value) && hasDefault {
		return defaultValue
	}
	return value
}

// hasTransitiveDefault finds defaults through transparent wrappers and references
// without materializing them early or changing the declaring node's coercion order.
func hasTransitiveDefault(schema Schema, depth int) bool {
	if schema == nil || depth >= maxValidationDepth {
		return false
	}
	if defaults, ok := schema.(interface{ defaultInfo() (any, bool) }); ok {
		if _, present := defaults.defaultInfo(); present {
			return true
		}
	}
	switch node := schema.(type) {
	case *OptionalSchema:
		return hasTransitiveDefault(node.inner, depth+1)
	case *NullableSchema:
		return hasTransitiveDefault(node.inner, depth+1)
	case *RefSchema:
		return hasTransitiveDefault(node.resolved, depth+1)
	}
	return false
}
