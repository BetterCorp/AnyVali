package anyvali

import (
	"fmt"
	"strconv"
	"strings"
)

// CoercionType represents a portable coercion kind.
type CoercionType string

const (
	CoerceToInt    CoercionType = "int"
	CoerceToNumber CoercionType = "number"
	CoerceToBool   CoercionType = "bool"
	CoerceTrim     CoercionType = "trim"
	CoerceLower    CoercionType = "lower"
	CoerceUpper    CoercionType = "upper"
)

// applyCoercion attempts to coerce the input value according to the given coercion type.
func applyCoercion(value any, coercion CoercionType) (any, error) {
	switch coercion {
	case CoerceToInt:
		return coerceToInt(value)
	case CoerceToNumber:
		return coerceToNumber(value)
	case CoerceToBool:
		return coerceToBool(value)
	case CoerceTrim:
		return coerceTrimString(value)
	case CoerceLower:
		return coerceLowerString(value)
	case CoerceUpper:
		return coerceUpperString(value)
	default:
		return nil, fmt.Errorf("unknown coercion type: %s", coercion)
	}
}

func coerceToInt(value any) (any, error) {
	switch v := value.(type) {
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot coerce string %q to int", v)
		}
		return i, nil
	default:
		return nil, fmt.Errorf("cannot coerce %T to int", value)
	}
}

func coerceToNumber(value any) (any, error) {
	switch v := value.(type) {
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot coerce string %q to number", v)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("cannot coerce %T to number", value)
	}
}

func coerceToBool(value any) (any, error) {
	switch v := value.(type) {
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("cannot coerce string %q to bool", v)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("cannot coerce %T to bool", value)
	}
}

func coerceTrimString(value any) (any, error) {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v), nil
	default:
		return nil, fmt.Errorf("cannot trim non-string value %T", value)
	}
}

func coerceLowerString(value any) (any, error) {
	switch v := value.(type) {
	case string:
		return strings.ToLower(v), nil
	default:
		return nil, fmt.Errorf("cannot lowercase non-string value %T", value)
	}
}

func coerceUpperString(value any) (any, error) {
	switch v := value.(type) {
	case string:
		return strings.ToUpper(v), nil
	default:
		return nil, fmt.Errorf("cannot uppercase non-string value %T", value)
	}
}
