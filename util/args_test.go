package util_test

import (
	"reflect"
	"testing"

	"github.com/qjebbs/go-sqls/util"
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
		got := util.Args(tc.slice)
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("want: %s, got: %s", tc.want, got)
		}
	}
}
