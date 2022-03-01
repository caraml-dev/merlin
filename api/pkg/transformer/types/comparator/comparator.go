package comparator

import (
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

type Comparator string

const (
	Greater   Comparator = ">"
	GreaterEq Comparator = ">="
	Less      Comparator = "<"
	LessEq    Comparator = "<="
	Eq        Comparator = "=="
	Neq       Comparator = "!="
	In        Comparator = "in"
)

func CompareInt64(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	compareFn := func(l, r int64, c Comparator) (bool, error) {
		switch c {
		case Greater:
			return (l > r), nil
		case GreaterEq:
			return (l >= r), nil
		case Less:
			return (l < r), nil
		case LessEq:
			return (l <= r), nil
		case Eq:
			return (l == r), nil
		case Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("")
		}
	}
	lValInt64, err := converter.ToInt64(lValue)
	if err != nil {
		return false, err
	}
	rValInt64, err := converter.ToInt64(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValInt64, rValInt64, comparator)
}

func CompareFloat64(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	compareFn := func(l, r float64, c Comparator) (bool, error) {
		switch c {
		case Greater:
			return (l > r), nil
		case GreaterEq:
			return (l >= r), nil
		case Less:
			return (l < r), nil
		case LessEq:
			return (l <= r), nil
		case Eq:
			return (l == r), nil
		case Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("%s operator is not supported", c)
		}
	}
	lValInt64, err := converter.ToFloat64(lValue)
	if err != nil {
		return false, err
	}
	rValInt64, err := converter.ToFloat64(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValInt64, rValInt64, comparator)
}

func CompareString(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	compareFn := func(l, r string, c Comparator) (bool, error) {
		switch c {
		case Greater:
			return (l > r), nil
		case GreaterEq:
			return (l >= r), nil
		case Less:
			return (l < r), nil
		case LessEq:
			return (l <= r), nil
		case Eq:
			return (l == r), nil
		case Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("%s operator is not supported", c)
		}
	}
	lValString, err := converter.ToString(lValue)
	if err != nil {
		return false, err
	}
	rValString, err := converter.ToString(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValString, rValString, comparator)
}

func CompareBool(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	compareFn := func(l, r bool, c Comparator) (bool, error) {
		switch c {
		case Eq:
			return (l == r), nil
		case Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("%s operator is not supported", c)
		}
	}
	lValBool, err := converter.ToBool(lValue)
	if err != nil {
		return false, err
	}
	rValBool, err := converter.ToBool(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValBool, rValBool, comparator)
}
