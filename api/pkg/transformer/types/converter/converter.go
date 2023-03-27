package converter

import (
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	feast "github.com/feast-dev/feast/sdk/go"
	feastType "github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	feastType2 "github.com/caraml-dev/merlin/pkg/transformer/types/feast"
)

func ToString(val interface{}) (string, error) {
	return fmt.Sprintf("%v", val), nil
}

func ToStringList(val interface{}) ([]string, error) {
	switch v := val.(type) {
	case string:
		return []string{v}, nil
	case []string:
		return v, nil
	default:
		switch reflect.TypeOf(val).Kind() {
		case reflect.Slice:
			v := reflect.ValueOf(val)
			l := v.Len()
			out := make([]string, l)
			for i := 0; i < l; i++ {
				out[i] = fmt.Sprintf("%v", v.Index(i))
			}
			return out, nil
		default:
			return []string{fmt.Sprintf("%v", val)}, nil
		}
	}
}

func ToInt(v interface{}) (int, error) {
	switch v := v.(type) {
	case *float64:
		return int(*v), nil
	case float64:
		return int(v), nil
	case *float32:
		return int(*v), nil
	case float32:
		return int(v), nil
	case *int:
		return int(*v), nil
	case int:
		return int(v), nil
	case *int8:
		return int(*v), nil
	case int8:
		return int(v), nil
	case *int16:
		return int(*v), nil
	case int16:
		return int(v), nil
	case *int32:
		return int(*v), nil
	case int32:
		return int(v), nil
	case *int64:
		return int(*v), nil
	case int64:
		return int(v), nil
	case *string:
		return strconv.Atoi(*v)
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to int", v)
	}
}

func ToIntList(val interface{}) ([]int, error) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v)
		return []int{int(i.Int())}, nil
	case []int, []int8, []int16, []int32, []int64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int, l)
		for i := 0; i < l; i++ {
			out[i] = int(list.Index(i).Int())
		}
		return out, nil
	case float32, float64:
		val := reflect.ValueOf(v)
		return []int{int(val.Float())}, nil
	case []float32, []float64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int, l)
		for i := 0; i < l; i++ {
			out[i] = int(list.Index(i).Float())
		}
		return out, nil
	case bool:
		if v {
			return []int{1}, nil
		} else {
			return []int{0}, nil
		}
	case []bool:
		l := len(v)
		out := make([]int, l)
		for i := 0; i < l; i++ {
			if v[i] {
				out[i] = 1
			} else {
				out[i] = 0
			}
		}
		return out, nil
	case string:
		d, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		return []int{d}, nil
	case []string:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int, l)
		for i := 0; i < l; i++ {
			d, err := strconv.Atoi(list.Index(i).String())
			if err != nil {
				return nil, err
			}
			out[i] = d
		}
		return out, nil
	case interface{}:
		switch reflect.TypeOf(val).Kind() {
		case reflect.Slice:
			list := reflect.ValueOf(v)
			l := list.Len()
			out := make([]int, l)
			for i := 0; i < l; i++ {
				d, err := ToInt(list.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				out[i] = d
			}
			return out, nil
		default:
			d, err := ToInt(v)
			if err != nil {
				return nil, err
			}
			return []int{d}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported conversion from %T to []int", v)
	}
}

func ToInt64(v interface{}) (int64, error) {
	switch v := v.(type) {
	case *float64:
		return int64(*v), nil
	case float64:
		return int64(v), nil
	case *float32:
		return int64(*v), nil
	case float32:
		return int64(v), nil
	case *int:
		return int64(*v), nil
	case int:
		return int64(v), nil
	case *int8:
		return int64(*v), nil
	case int8:
		return int64(v), nil
	case *int16:
		return int64(*v), nil
	case int16:
		return int64(v), nil
	case *int32:
		return int64(*v), nil
	case int32:
		return int64(v), nil
	case *int64:
		return int64(*v), nil
	case int64:
		return int64(v), nil
	case *string:
		return strconv.ParseInt(*v, 10, 64)
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to int", v)
	}
}

func ToInt32(v interface{}) (int32, error) {
	switch v := v.(type) {
	case *float64:
		return int32(*v), nil
	case float64:
		return int32(v), nil
	case *float32:
		return int32(*v), nil
	case float32:
		return int32(v), nil
	case *int:
		return int32(*v), nil
	case int:
		return int32(v), nil
	case *int8:
		return int32(*v), nil
	case int8:
		return int32(v), nil
	case *int16:
		return int32(*v), nil
	case int16:
		return int32(v), nil
	case *int32:
		return *v, nil
	case int32:
		return v, nil
	case *int64:
		return int32(*v), nil
	case int64:
		return int32(v), nil
	case *string:
		val, err := strconv.ParseInt(*v, 10, 64)
		if err != nil {
			return 0, err
		}
		return int32(val), nil
	case string:
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return int32(val), nil
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to int", v)
	}
}

func ToInt32List(val interface{}) ([]int32, error) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v)
		return []int32{int32(i.Int())}, nil
	case []int, []int8, []int16, []int32, []int64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int32, l)
		for i := 0; i < l; i++ {
			out[i] = int32(list.Index(i).Int())
		}
		return out, nil
	case float32, float64:
		val := reflect.ValueOf(v)
		return []int32{int32(val.Float())}, nil
	case []float32, []float64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int32, l)
		for i := 0; i < l; i++ {
			out[i] = int32(list.Index(i).Float())
		}
		return out, nil
	case bool:
		if v {
			return []int32{1}, nil
		} else {
			return []int32{0}, nil
		}
	case []bool:
		l := len(v)
		out := make([]int32, l)
		for i := 0; i < l; i++ {
			if v[i] {
				out[i] = 1
			} else {
				out[i] = 0
			}
		}
		return out, nil
	case string:
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		return []int32{int32(d)}, nil
	case []string:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int32, l)
		for i := 0; i < l; i++ {
			d, err := strconv.ParseInt(list.Index(i).String(), 10, 64)
			if err != nil {
				return nil, err
			}
			out[i] = int32(d)
		}
		return out, nil
	case interface{}:
		switch reflect.TypeOf(val).Kind() {
		case reflect.Slice:
			list := reflect.ValueOf(v)
			l := list.Len()
			out := make([]int32, l)
			for i := 0; i < l; i++ {
				d, err := ToInt32(list.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				out[i] = d
			}
			return out, nil
		default:
			d, err := ToInt32(v)
			if err != nil {
				return nil, err
			}
			return []int32{d}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported conversion from %T to []int32", v)
	}
}

func ToInt64List(val interface{}) ([]int64, error) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v)
		return []int64{int64(i.Int())}, nil
	case []int, []int8, []int16, []int32, []int64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int64, l)
		for i := 0; i < l; i++ {
			out[i] = list.Index(i).Int()
		}
		return out, nil
	case float32, float64:
		val := reflect.ValueOf(v)
		return []int64{int64(val.Float())}, nil
	case []float32, []float64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int64, l)
		for i := 0; i < l; i++ {
			out[i] = int64(list.Index(i).Float())
		}
		return out, nil
	case bool:
		if v {
			return []int64{1}, nil
		} else {
			return []int64{0}, nil
		}
	case []bool:
		l := len(v)
		out := make([]int64, l)
		for i := 0; i < l; i++ {
			if v[i] {
				out[i] = 1
			} else {
				out[i] = 0
			}
		}
		return out, nil
	case string:
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		return []int64{d}, nil
	case []string:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]int64, l)
		for i := 0; i < l; i++ {
			d, err := strconv.ParseInt(list.Index(i).String(), 10, 64)
			if err != nil {
				return nil, err
			}
			out[i] = d
		}
		return out, nil
	case interface{}:
		switch reflect.TypeOf(val).Kind() {
		case reflect.Slice:
			list := reflect.ValueOf(v)
			l := list.Len()
			out := make([]int64, l)
			for i := 0; i < l; i++ {
				d, err := ToInt64(list.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				out[i] = d
			}
			return out, nil
		default:
			d, err := ToInt64(v)
			if err != nil {
				return nil, err
			}
			return []int64{d}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported conversion from %T to []int64", v)
	}
}

func ToFloat32(v interface{}) (float32, error) {
	switch v := v.(type) {
	case float64:
		return float32(v), nil
	case float32:
		return v, nil
	case int:
		return float32(v), nil
	case int8:
		return float32(v), nil
	case int16:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case string:
		floatVal, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return 0, err
		}
		return float32(floatVal), nil
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to float32", v)
	}
}

func ToFloat64(v interface{}) (float64, error) {
	switch v := v.(type) {
	case *float64:
		return *v, nil
	case float64:
		return v, nil
	case *float32:
		return float64(*v), nil
	case float32:
		return float64(v), nil
	case *int:
		return float64(*v), nil
	case int:
		return float64(v), nil
	case *int8:
		return float64(*v), nil
	case int8:
		return float64(v), nil
	case *int16:
		return float64(*v), nil
	case int16:
		return float64(v), nil
	case *int32:
		return float64(*v), nil
	case int32:
		return float64(v), nil
	case *int64:
		return float64(*v), nil
	case int64:
		return float64(v), nil
	case *string:
		return strconv.ParseFloat(*v, 64)
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to float64", v)
	}
}

func ToFloat32List(val interface{}) ([]float32, error) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v)
		return []float32{float32(i.Int())}, nil
	case []int, []int8, []int16, []int32, []int64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]float32, l)
		for i := 0; i < l; i++ {
			out[i] = float32(list.Index(i).Int())
		}
		return out, nil
	case float32, float64:
		d, err := ToFloat32(val)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(float64(d)) {
			return []float32{}, nil
		}
		return []float32{d}, nil
	case []float32, []float64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]float32, 0, l)
		for i := 0; i < l; i++ {
			float64Val := list.Index(i).Float()
			if math.IsNaN(float64Val) {
				continue
			}
			out = append(out, float32(float64Val))
		}
		return out, nil
	case bool:
		if v {
			return []float32{1}, nil
		} else {
			return []float32{0}, nil
		}
	case []bool:
		l := len(v)
		out := make([]float32, l)
		for i := 0; i < l; i++ {
			if v[i] {
				out[i] = 1
			} else {
				out[i] = 0
			}
		}
		return out, nil
	case string:
		d, err := ToFloat32(val)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(float64(d)) {
			return []float32{}, nil
		}
		return []float32{float32(d)}, nil
	case []string:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]float32, 0, l)
		for i := 0; i < l; i++ {
			d, err := ToFloat32(list.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			if math.IsNaN(float64(d)) {
				continue
			}
			out = append(out, d)
		}
		return out, nil
	case interface{}:
		switch reflect.TypeOf(val).Kind() {
		case reflect.Slice:
			list := reflect.ValueOf(v)
			l := list.Len()
			out := make([]float32, 0, l)
			for i := 0; i < l; i++ {
				d, err := ToFloat32(list.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				if math.IsNaN(float64(d)) {
					continue
				}
				out = append(out, d)
			}
			return out, nil
		default:
			d, err := ToFloat32(v)
			if err != nil {
				return nil, err
			}
			if math.IsNaN(float64(d)) {
				return []float32{}, nil
			}
			return []float32{d}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported conversion from %T to []float32", v)
	}
}

func ToFloat64List(val interface{}) ([]float64, error) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v)
		return []float64{float64(i.Int())}, nil
	case []int, []int8, []int16, []int32, []int64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]float64, l)
		for i := 0; i < l; i++ {
			float64Val := float64(list.Index(i).Int())
			out[i] = float64Val
		}
		return out, nil
	case float32, float64:
		float64Val, err := ToFloat64(val)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(float64Val) {
			return []float64{}, nil
		}
		return []float64{float64Val}, nil
	case []float32, []float64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]float64, 0, l)
		for i := 0; i < l; i++ {
			float64Val := list.Index(i).Float()
			if math.IsNaN(float64Val) {
				continue
			}
			out = append(out, float64Val)
		}
		return out, nil
	case bool:
		if v {
			return []float64{1}, nil
		} else {
			return []float64{0}, nil
		}
	case []bool:
		l := len(v)
		out := make([]float64, l)
		for i := 0; i < l; i++ {
			if v[i] {
				out[i] = 1
			} else {
				out[i] = 0
			}
		}
		return out, nil
	case string:
		d, err := ToFloat64(val)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(d) {
			return []float64{}, nil
		}
		return []float64{d}, nil
	case []string:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]float64, 0, l)
		for i := 0; i < l; i++ {
			d, err := ToFloat64(list.Index(i).String())
			if err != nil {
				return nil, err
			}
			if math.IsNaN(d) {
				continue
			}
			out = append(out, d)
		}
		return out, nil
	case interface{}:
		switch reflect.TypeOf(val).Kind() {
		case reflect.Slice:
			list := reflect.ValueOf(v)
			l := list.Len()
			out := make([]float64, 0, l)
			for i := 0; i < l; i++ {
				d, err := ToFloat64(list.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				if math.IsNaN(d) {
					continue
				}
				out = append(out, d)
			}
			return out, nil
		default:
			d, err := ToFloat64(v)
			if err != nil {
				return nil, err
			}
			if math.IsNaN(d) {
				return []float64{}, nil
			}

			return []float64{d}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported conversion from %T to []float64", v)
	}
}

func ToBool(v interface{}) (bool, error) {
	switch v := v.(type) {
	case *bool:
		return *v, nil
	case bool:
		return v, nil
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i == 1 {
			return true, nil
		} else if i == 0 {
			return false, nil
		}
		return false, fmt.Errorf("error parsing %v to bool", v)
	case float32, float64:
		i := reflect.ValueOf(v).Float()
		if i == float64(1) {
			return true, nil
		} else if i == float64(0) {
			return false, nil
		}
		return false, fmt.Errorf("error parsing %v to bool", v)
	case *string:
		return strconv.ParseBool(*v)
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("unsupported conversion from %T to bool", v)
	}
}

func ToBoolList(val interface{}) ([]bool, error) {
	switch v := val.(type) {
	case bool:
		return []bool{v}, nil
	case []bool:
		return v, nil
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v)
		if i.Int() == 1 {
			return []bool{true}, nil
		} else if i.Int() == 0 {
			return []bool{false}, nil
		}
		return nil, fmt.Errorf("error parsing %v to bool", v)
	case []int, []int8, []int16, []int32, []int64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]bool, l)
		for i := 0; i < l; i++ {
			if list.Index(i).Int() == 1 {
				out[i] = true
			} else if list.Index(i).Int() == 0 {
				out[i] = false
			} else {
				return nil, fmt.Errorf("error parsing %v to bool", v)
			}
		}
		return out, nil
	case float32, float64:
		i := reflect.ValueOf(v)
		if i.Float() == 1 {
			return []bool{true}, nil
		} else if i.Float() == 0 {
			return []bool{false}, nil
		} else {
			return nil, fmt.Errorf("error parsing %v to bool", v)
		}
	case []float32, []float64:
		list := reflect.ValueOf(v)
		l := list.Len()
		out := make([]bool, l)
		for i := 0; i < l; i++ {
			if list.Index(i).Float() == 1 {
				out[i] = true
			} else if list.Index(i).Float() == 0 {
				out[i] = false
			} else {
				return nil, fmt.Errorf("error parsing %v to bool", v)
			}
		}
		return out, nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		return []bool{b}, nil
	case []string:
		l := len(v)
		out := make([]bool, l)
		for i := 0; i < l; i++ {
			b, err := strconv.ParseBool(v[i])
			if err != nil {
				return nil, err
			}
			out[i] = b
		}
		return out, nil
	case interface{}:
		switch reflect.TypeOf(val).Kind() {
		case reflect.Slice:
			list := reflect.ValueOf(v)
			l := list.Len()
			out := make([]bool, l)
			for i := 0; i < l; i++ {
				d, err := ToBool(list.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				out[i] = d
			}
			return out, nil
		default:
			d, err := ToBool(v)
			if err != nil {
				return nil, err
			}
			return []bool{d}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported conversion from %T to []bool", v)
	}
}

func ToTargetType(val interface{}, targetType spec.ValueType) (interface{}, error) {
	var value interface{}
	var err error
	switch targetType {
	case spec.ValueType_STRING:
		value, err = ToString(val)
	case spec.ValueType_INT:
		value, err = ToInt(val)
	case spec.ValueType_FLOAT:
		value, err = ToFloat64(val)
	case spec.ValueType_BOOL:
		value, err = ToBool(val)
	case spec.ValueType_STRING_LIST:
		value, err = ToStringList(val)
	case spec.ValueType_INT_LIST:
		value, err = ToIntList(val)
	case spec.ValueType_FLOAT_LIST:
		value, err = ToFloat64List(val)
	case spec.ValueType_BOOL_LIST:
		value, err = ToBoolList(val)
	default:
		return nil, fmt.Errorf("targetType is not recognized %d", targetType)
	}
	return value, err
}

func ToFeastValue(v interface{}, valueType feastType.ValueType_Enum) (*feastType.Value, error) {
	switch valueType {
	case feastType.ValueType_INT32:
		val, err := ToInt(v)
		if err != nil {
			return nil, err
		}
		return feast.Int32Val(int32(val)), nil
	case feastType.ValueType_INT64:
		val, err := ToInt64(v)
		if err != nil {
			return nil, err
		}
		return feast.Int64Val(val), nil
	case feastType.ValueType_FLOAT:
		val, err := ToFloat64(v)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(val) {
			return nil, nil
		}
		return feast.FloatVal(float32(val)), nil
	case feastType.ValueType_DOUBLE:
		val, err := ToFloat64(v)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(val) {
			return nil, nil
		}
		return feast.DoubleVal(val), nil
	case feastType.ValueType_BOOL:
		val, err := ToBool(v)
		if err != nil {
			return nil, err
		}
		return feast.BoolVal(val), nil
	case feastType.ValueType_STRING:
		switch inType := v.(type) {
		case float64:
			// we'll truncate decimal point as number in json is treated as float64 and it doesn't make sense to have decimal as entity id
			return feast.StrVal(fmt.Sprintf("%.0f", inType)), nil
		default:
			val, err := ToString(v)
			if err != nil {
				return nil, err
			}
			return feast.StrVal(val), nil
		}
	case feastType.ValueType_INT32_LIST:
		switch v.(type) {
		case string:
			arr, err := extractArrayFromString(v, ",")
			if err != nil {
				return nil, err
			}
			v = arr
		}
		val, err := ToInt32List(v)
		if err != nil {
			return nil, err
		}
		return feastType2.Int32ListVal(val), nil
	case feastType.ValueType_INT64_LIST:
		switch v.(type) {
		case string:
			arr, err := extractArrayFromString(v, ",")
			if err != nil {
				return nil, err
			}
			v = arr
		}
		val, err := ToInt64List(v)
		if err != nil {
			return nil, err
		}
		return feastType2.Int64ListVal(val), nil
	case feastType.ValueType_FLOAT_LIST:
		switch v.(type) {
		case string:
			arr, err := extractArrayFromString(v, ",")
			if err != nil {
				return nil, err
			}
			v = arr
		}
		val, err := ToFloat32List(v)
		if err != nil {
			return nil, err
		}
		return feastType2.FloatListVal(val), nil
	case feastType.ValueType_DOUBLE_LIST:
		switch v.(type) {
		case string:
			arr, err := extractArrayFromString(v, ",")
			if err != nil {
				return nil, err
			}
			v = arr
		}
		val, err := ToFloat64List(v)
		if err != nil {
			return nil, err
		}
		return feastType2.DoubleListVal(val), nil
	case feastType.ValueType_BOOL_LIST:
		switch v.(type) {
		case string:
			arr, err := extractArrayFromString(v, ",")
			if err != nil {
				return nil, err
			}
			v = arr
		}
		val, err := ToBoolList(v)
		if err != nil {
			return nil, err
		}
		return feastType2.BoolListVal(val), nil
	case feastType.ValueType_STRING_LIST:
		switch v.(type) {
		case string:
			arr, err := extractArrayFromString(v, ",")
			if err != nil {
				return nil, err
			}
			v = arr
		}
		val, err := ToStringList(v)
		if err != nil {
			return nil, err
		}
		return feastType2.StrListVal(val), nil
	default:
		return nil, fmt.Errorf("unsupported type %s", valueType.String())
	}
}

func extractArrayFromString(str interface{}, delimiter string) (interface{}, error) {
	val, err := ToString(str)
	if err != nil {
		return nil, err
	}
	l := len(val)
	v := val
	if val[0] == '[' && val[l-1] == ']' {
		v = val[1 : l-1]
	}
	v = strings.ReplaceAll(v, " ", "")
	return strings.Split(v, delimiter), nil
}

func ExtractFeastValue(val *feastType.Value) (interface{}, feastType.ValueType_Enum, error) {
	switch val.Val.(type) {
	case *feastType.Value_StringVal:
		return val.GetStringVal(), feastType.ValueType_STRING, nil
	case *feastType.Value_DoubleVal:
		return val.GetDoubleVal(), feastType.ValueType_DOUBLE, nil
	case *feastType.Value_FloatVal:
		return val.GetFloatVal(), feastType.ValueType_FLOAT, nil
	case *feastType.Value_Int32Val:
		return val.GetInt32Val(), feastType.ValueType_INT32, nil
	case *feastType.Value_Int64Val:
		return val.GetInt64Val(), feastType.ValueType_INT64, nil
	case *feastType.Value_BoolVal:
		return val.GetBoolVal(), feastType.ValueType_BOOL, nil
	case *feastType.Value_StringListVal:
		return val.GetStringListVal().GetVal(), feastType.ValueType_STRING_LIST, nil
	case *feastType.Value_DoubleListVal:
		return val.GetDoubleListVal().GetVal(), feastType.ValueType_DOUBLE_LIST, nil
	case *feastType.Value_FloatListVal:
		return val.GetFloatListVal().GetVal(), feastType.ValueType_FLOAT_LIST, nil
	case *feastType.Value_Int32ListVal:
		return val.GetInt32ListVal().GetVal(), feastType.ValueType_INT32_LIST, nil
	case *feastType.Value_Int64ListVal:
		return val.GetInt64ListVal().GetVal(), feastType.ValueType_INT64_LIST, nil
	case *feastType.Value_BoolListVal:
		return val.GetBoolListVal().GetVal(), feastType.ValueType_BOOL_LIST, nil
	case *feastType.Value_BytesVal:
		return base64.StdEncoding.EncodeToString(val.GetBytesVal()), feastType.ValueType_STRING, nil
	case *feastType.Value_BytesListVal:
		results := make([]string, 0)
		for _, bytes := range val.GetBytesListVal().GetVal() {
			results = append(results, base64.StdEncoding.EncodeToString(bytes))
		}
		return results, feastType.ValueType_STRING_LIST, nil
	default:
		return nil, feastType.ValueType_INVALID, fmt.Errorf("unknown feature cacheValue type: %T", val.Val)
	}
}
