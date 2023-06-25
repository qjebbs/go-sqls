package util

import (
	"reflect"
)

// Args is a help func to convert s to query args.
// if s is not a slice or array, it will return []any{s}.
func Args(s any) []any {
	switch a := s.(type) {
	case []any:
		return a
	case []bool:
		return Ttoa(a)
	case []float64:
		return Ttoa(a)
	case []float32:
		return Ttoa(a)
	case []int64:
		return Ttoa(a)
	case []int32:
		return Ttoa(a)
	case []string:
		return Ttoa(a)
	case *[]bool:
		return Ttoa(*a)
	case *[]float64:
		return Ttoa(*a)
	case *[]float32:
		return Ttoa(*a)
	case *[]int64:
		return Ttoa(*a)
	case *[]int32:
		return Ttoa(*a)
	case *[]string:
		return Ttoa(*a)
	default:
		return convertArrayReflect(s)
	}
}

// Ttoa is a help func to convert slice to []any.
func Ttoa[T any](slice []T) []any {
	if len(slice) == 0 {
		return nil
	}
	b := make([]any, 0, len(slice))
	for _, v := range slice {
		b = append(b, v)
	}
	return b
}

func convertArrayReflect(slice any) []any {
	rv := reflect.ValueOf(slice)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return nil
		}
	case reflect.Array:
	default:
		return []any{slice}
	}
	n := rv.Len()
	if n == 0 {
		return nil
	}
	s := make([]any, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, rv.Index(i).Interface())
	}
	return s
}
