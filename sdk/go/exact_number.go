package anyvali

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
)

// exactNumberRat compares numeric JSON values without rounding wide integers
// through float64. Bound parsing so a hostile exponent cannot allocate an
// arbitrarily large big.Rat during validation.
func exactNumberRat(value any) (*big.Rat, bool) {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
		float32, float64, json.Number:
	default:
		return nil, false
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}
	number := string(data)
	mantissa := number
	if index := strings.IndexAny(number, "eE"); index >= 0 {
		mantissa = number[:index]
	}
	// JSON spelling and exponent size do not change zero's value. Recognize it
	// before bounded rational parsing, without expanding a hostile exponent.
	if strings.Trim(mantissa, "-0.") == "" {
		return new(big.Rat), true
	}
	if len(data) > 128 {
		return nil, false
	}
	exponent := 0
	if index := strings.IndexAny(number, "eE"); index >= 0 {
		exponent, err = strconv.Atoi(number[index+1:])
		if err != nil || exponent < -1000 || exponent > 1000 {
			return nil, false
		}
		number = number[:index]
	}
	rat, ok := new(big.Rat).SetString(number)
	if !ok {
		return nil, false
	}
	if exponent > 0 {
		rat.Mul(rat, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)))
	} else if exponent < 0 {
		rat.Quo(rat, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-exponent)), nil)))
	}
	return rat, ok
}
