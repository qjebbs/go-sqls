package sqls_test

import (
	"reflect"
	"testing"

	"git.qjebbs.com/jebbs/go-sqls"
)

func TestBuildSegment(t *testing.T) {
	t.Parallel()
	table := sqls.Table{"table", "t"}
	testCases := []struct {
		segment  *sqls.Segment
		want     string
		wantArgs []any
		wantErr  bool
	}{
		{
			segment:  nil,
			want:     "",
			wantArgs: []any{},
		},
		{
			segment: &sqls.Segment{
				Raw:  "#join('#?',','),#?(1),#?(2)",
				Args: []any{1, 2},
			},
			want:     "?,?,?,?",
			wantArgs: []any{1, 2, 1, 2},
		},
		{
			segment: &sqls.Segment{
				Raw:      "WHERE 1=1 #s1",
				Segments: []*sqls.Segment{nil},
			},
			want:     "WHERE 1=1",
			wantArgs: []any{},
		},
		{
			segment: &sqls.Segment{
				Raw:     "WHERE #c1=?",
				Columns: table.Columns("id"),
				Args:    []any{nil},
			},
			want:     "WHERE t.id=?",
			wantArgs: []any{nil},
		},
		{
			segment: &sqls.Segment{
				Raw:     "WHERE #c1=$1",
				Columns: []*sqls.TableColumn{nil},
				Args:    []any{nil},
			},
			want:     "WHERE =$1",
			wantArgs: []any{nil},
		},
		{
			segment: &sqls.Segment{
				Raw:     "#c1>1",
				Columns: table.Columns("id"),
				Args:    nil,
			},
			want:     "t.id>1",
			wantArgs: []any{},
		},
		{
			segment: &sqls.Segment{
				Raw:     "#c2 IS NULL AND #c1>$1",
				Columns: table.Columns("id", "deleted"),
				Args:    []any{1},
			},
			want:     "t.deleted IS NULL AND t.id>$1",
			wantArgs: []any{1},
		},
		{
			segment: &sqls.Segment{
				Raw:     "#c1>$1",
				Columns: table.Columns("id"),
				Args:    []any{1},
			},
			want:     "t.id>$1",
			wantArgs: []any{1},
		},
		{
			segment: &sqls.Segment{
				Raw:     "#c1 IN ($2,$1)",
				Columns: table.Columns("id"),
				Args:    []any{1, 2},
			},
			want:     "t.id IN ($1,$2)",
			wantArgs: []any{2, 1},
		},
		{
			segment: &sqls.Segment{
				Raw: "#c1",
				Columns: []*sqls.TableColumn{
					table.Expression("#t1.id=$1", 1),
				},
			},
			want:     "t.id=$1",
			wantArgs: []any{1},
		},
		{
			segment: &sqls.Segment{
				Raw: "#c1 > $1",
				Columns: []*sqls.TableColumn{
					table.Expression("#t1.id - $1", 1),
				},
				Args: []any{2},
			},
			want:     "t.id - $1 > $2",
			wantArgs: []any{1, 2},
		},
		{
			segment: &sqls.Segment{
				Raw: "WITH t AS (#s1) SELECT #c1,#c2,$1 FROM #tAs1",
				Segments: []*sqls.Segment{
					{
						Raw:     "SELECT * FROM #tAs1 WHERE #c1 > $1",
						Columns: table.Columns("id"),
						Args:    []any{1},
					},
				},
				Columns: []*sqls.TableColumn{
					table.Expression("#t1.id"),
					table.Expression("#t1.id=$1", 2),
				},
				Args: []any{"foo"},
			},
			want:     "WITH t AS (SELECT * FROM table AS t WHERE t.id > $1) SELECT t.id,t.id=$2,$3 FROM table AS t",
			wantArgs: []any{1, 2, "foo"},
		},
		{
			segment: &sqls.Segment{
				Raw: "SELECT #join('#c', ', ') FROM #tAs1",
				Columns: []*sqls.TableColumn{
					table.Expression("#t1.id"),
					table.Expression("#t1.id=$1", 1),
					table.Expression("#t1.name"),
				},
			},
			want:     "SELECT t.id, t.id=$1, t.name FROM table AS t",
			wantArgs: []any{1},
		},
		{
			segment: &sqls.Segment{
				Raw:      "#s1",
				Segments: []*sqls.Segment{nil},
				Header:   "WHERE",
				Footer:   "FOR UPDATE",
			},
			want:     "",
			wantArgs: []any{},
		},
		{
			segment: &sqls.Segment{
				Raw: "#s1",
				Segments: []*sqls.Segment{
					{
						Raw:     "#c1=$1",
						Columns: table.Columns("id"),
						Args:    []any{1},
					},
				},
				Header: "WHERE",
				Footer: "FOR UPDATE",
			},
			want:     "WHERE t.id=$1 FOR UPDATE",
			wantArgs: []any{1},
		},
		{
			segment: &sqls.Segment{
				Raw: "#s1, #s1",
				Segments: []*sqls.Segment{{
					Raw:  "#join('#?', ', '), ?",
					Args: []any{1, 2},
				}},
			},
			want:     "?, ?, ?, ?, ?, ?",
			wantArgs: []any{1, 2, 1, 1, 2, 1},
		},
		{
			segment: &sqls.Segment{
				Raw:  "?, $1",
				Args: []any{nil},
			},
			wantErr: true,
		},
		{
			segment: &sqls.Segment{
				Raw: "? #s1",
				Segments: []*sqls.Segment{{
					Raw:  "$1",
					Args: []any{2},
				}},
				Args: []any{1},
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		name := "nil"
		if tc.segment != nil {
			name = tc.segment.Raw
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			args := make([]any, 0)
			got, err := tc.segment.BuildTo(&args)
			if err != nil {
				if tc.wantErr {
					return
				}
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Errorf("got %v, want %v", args, tc.wantArgs)
			}
		})
	}
}
