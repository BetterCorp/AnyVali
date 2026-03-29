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
