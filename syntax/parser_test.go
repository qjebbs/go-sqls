package syntax_test

import (
	"testing"

	"git.qjebbs.com/jebbs/go-sqls/syntax"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		raw     string
		want    []syntax.Expr
		wantErr bool
	}{
		{
			raw:     "?1",
			wantErr: true,
		},
		{
			raw:     "$",
			wantErr: true,
		},
		{
			raw:     "$1,?",
			wantErr: true,
		},
		{
			raw: "?,?,?",
			want: []syntax.Expr{
				&syntax.RefExpr{Type: syntax.ArgUnindexed, Index: 1},
				&syntax.PlainExpr{Text: ","},
				&syntax.RefExpr{Type: syntax.ArgUnindexed, Index: 2},
				&syntax.PlainExpr{Text: ","},
				&syntax.RefExpr{Type: syntax.ArgUnindexed, Index: 3},
			},
		},
		{
			raw: "$1'#c11#t111#s1111'",
			want: []syntax.Expr{
				&syntax.RefExpr{Type: syntax.ArgIndexed, Index: 1},
				&syntax.PlainExpr{Text: "'#c11#t111#s1111'"},
			},
		},
		{
			raw: "#join('#c=#$', ',')",
			want: []syntax.Expr{
				&syntax.FuncCallExpr{
					Name: "join",
					Args: []string{"#c=#$", ","},
				},
			},
		},
		{
			raw: "#c1#t1#s1",
			want: []syntax.Expr{
				&syntax.FuncCallExpr{Name: "c", Args: []string{"1"}},
				&syntax.FuncCallExpr{Name: "t", Args: []string{"1"}},
				&syntax.FuncCallExpr{Name: "s", Args: []string{"1"}},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.raw, func(t *testing.T) {
			got, err := syntax.Parse(tc.raw)
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}
			if !tc.wantErr && !cmp.Equal(
				got.ExprList, tc.want,
				cmpopts.IgnoreUnexported(
					syntax.PlainExpr{},
					syntax.RefExpr{},
					syntax.FuncExpr{},
					syntax.FuncCallExpr{},
				),
			) {
				for _, tk := range got.ExprList {
					t.Logf("%#v", tk)
				}
				t.Fatal("failed")
			}
		})
	}
}
