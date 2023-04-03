package slices_test

import (
	"reflect"
	"testing"

	"git.qjebbs.com/jebbs/go-sqls/slices"
)

func TestConvertSlices(t *testing.T) {
	type str string
	strA := str("a")
	testCases := []struct {
		slice any
		want  []any
	}{
		{slice: []int{1, 2, 3}, want: []any{1, 2, 3}},
		{slice: []string{"a", "b", "c"}, want: []any{"a", "b", "c"}},
		{slice: []str{"a", "b", "c"}, want: []any{str("a"), str("b"), str("c")}},
		{slice: []*str{&strA}, want: []any{&strA}},
		{slice: []any{1, "a", 2, "b", 3, "c"}, want: []any{1, "a", 2, "b", 3, "c"}},
		{slice: 1, want: []any{1}},
	}
	for _, tc := range testCases {
		got := slices.Ttoa(tc.slice)
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("want: %s, got: %s", tc.want, got)
		}
	}
}
