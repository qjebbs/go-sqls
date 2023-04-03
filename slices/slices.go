package slices

import (
	"reflect"
)

// Concat concatenates multiple slices into one.
func Concat[T any](slices ...[]T) []T {
	switch len(slices) {
	case 0:
		return nil
	case 1:
		return slices[0]
	}
	r := make([]T, 0)
	for _, s := range slices {
		r = append(r, s...)
	}
	return r
}

// Atot (any to T) converts a slice []any to a slice of []T.
func Atot[T any](items []any) []T {
	slice := make([]T, 0, len(items))
	for _, item := range items {
		slice = append(slice, item.(T))
	}
	return slice
}

// Ttoa (T to any) is a help func to convert slice other types
// to []any, if s is not a slice or array, it will return []any{s}.
func Ttoa(s any) []any {
	switch a := s.(type) {
	case []any:
		return a
	case []bool:
		return convertSliceT(a)
	case []float64:
		return convertSliceT(a)
	case []float32:
		return convertSliceT(a)
	case []int64:
		return convertSliceT(a)
	case []int32:
		return convertSliceT(a)
	case []string:
		return convertSliceT(a)
	case *[]bool:
		return convertSliceT(*a)
	case *[]float64:
		return convertSliceT(*a)
	case *[]float32:
		return convertSliceT(*a)
	case *[]int64:
		return convertSliceT(*a)
	case *[]int32:
		return convertSliceT(*a)
	case *[]string:
		return convertSliceT(*a)
	default:
		return convertArrayReflect(s)
	}
}

func convertSliceT[T any](slice []T) []any {
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
