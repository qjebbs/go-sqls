package sqls

import (
	"testing"
)

func TestColumns(t *testing.T) {
	table := Table{"foo", "f"}
	testCases := []struct {
		column *TableColumn
		want   string
	}{
		{column: table.Column("id"), want: "f.id"},
		{column: table.Expression("id"), want: "id"},
		{column: table.Expression("COALESCE(#t1.id,0)"), want: "COALESCE(f.id,0)"},
		{column: table.Expression("COALESCE(#t1.id,$1)", 1), want: "COALESCE(f.id,$1)"},
	}
	for _, tc := range testCases {
		argStore := []any{}
		got, err := tc.column.buildInternal(newContext(&argStore, &Segment{
			Args: tc.column.Args,
		}))
		if err != nil {
			t.Errorf("want: %s, got: %s", tc.want, err)
			continue
		}
		if tc.want != got {
			t.Errorf("want: %s, got: %s", tc.want, got)
		}
	}
}
