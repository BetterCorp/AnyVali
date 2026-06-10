package anyvali

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Strict ASCII decimal grammars. Go's strconv parsers are more permissive than
// the ECMA-262 reference (JS): ParseInt accepts a leading "+", ParseFloat
// accepts hex floats ("0x1p4"), digit-group underscores ("1_000") and
// "inf"/"nan", and ParseBool accepts "t"/"T"/"f"/"F". Each let a string that
// every other SDK rejects coerce into a number/bool -- a cross-language
// validation bypass. Gate on these before parsing (spec 5.1: decimal only).
var (
	decimalIntRe   = regexp.MustCompile(`^-?[0-9]+$`)
	decimalFloatRe = regexp.MustCompile(`^[+-]?([0-9]+\.?[0-9]*|\.[0-9]+)([eE][+-]?[0-9]+)?$`)
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
		trimmed := strings.TrimSpace(v)
		if !decimalIntRe.MatchString(trimmed) {
			return nil, fmt.Errorf("cannot coerce string %q to int", v)
		}
		i, err := strconv.ParseInt(trimmed, 10, 64)
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
		trimmed := strings.TrimSpace(v)
		if !decimalFloatRe.MatchString(trimmed) {
			return nil, fmt.Errorf("cannot coerce string %q to number", v)
		}
		f, err := strconv.ParseFloat(trimmed, 64)
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
		// Spec 5.1: only "true"/"1" and "false"/"0" (case-insensitive).
		// strconv.ParseBool also accepts "t"/"T"/"f"/"F", which diverge from
		// the JS reference.
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1":
			return true, nil
		case "false", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("cannot coerce string %q to bool", v)
		}
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
