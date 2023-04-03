package syntax

import (
	"reflect"
	"testing"
)

func TestScanner(t *testing.T) {
	testCases := []struct {
		raw  string
		want []token
	}{
		{
			raw: "$1,$2",
			want: []token{
				{typ: _Ref, lit: "$", bad: false, kind: _StringLit, start: 0, end: 1},
				{typ: _Literal, lit: "1", bad: false, kind: _IntLit, start: 1, end: 2},
				{typ: _Plain, lit: ",", bad: false, kind: _StringLit, start: 2, end: 3},
				{typ: _Ref, lit: "$", bad: false, kind: _StringLit, start: 3, end: 4},
				{typ: _Literal, lit: "2", bad: false, kind: _IntLit, start: 4, end: 5},
				{typ: _EOF, lit: "", bad: false, kind: _StringLit, start: 5, end: 5},
			},
		},
		{
			raw: "a IN (?,?)",
			want: []token{
				{typ: _Plain, lit: "a IN (", bad: false, kind: _StringLit, start: 0, end: 6},
				{typ: _Ref, lit: "?", bad: false, kind: _StringLit, start: 6, end: 7},
				{typ: _Plain, lit: ",", bad: false, kind: _StringLit, start: 7, end: 8},
				{typ: _Ref, lit: "?", bad: false, kind: _StringLit, start: 8, end: 9},
				{typ: _Plain, lit: ")", bad: false, kind: _StringLit, start: 9, end: 10},
				{typ: _EOF, lit: "", bad: false, kind: _StringLit, start: 10, end: 10},
			},
		},
		{
			raw: "'a''b'",
			want: []token{
				{typ: _Plain, lit: "'a''b'", bad: false, kind: _StringLit, start: 0, end: 6},
				{typ: _EOF, lit: "", bad: false, kind: _StringLit, start: 6, end: 6},
			},
		},
		{
			raw: "'a''b",
			want: []token{
				{typ: _Plain, lit: "'a''b", bad: true, kind: _StringLit, start: 0, end: 5},
				{typ: _EOF, lit: "", bad: false, kind: _StringLit, start: 5, end: 5},
			},
		},
		{
			raw: "$$1",
			want: []token{
				{typ: _Plain, lit: "$$1", bad: false, kind: _StringLit, start: 0, end: 3},
				{typ: _EOF, lit: "", bad: false, kind: _StringLit, start: 3, end: 3},
			},
		},
		{
			raw: "#a(1,2)aaaa",
			want: []token{
				{typ: _Hash, lit: "#", bad: false, kind: _StringLit, start: 0, end: 1},
				{typ: _Name, lit: "a", bad: false, kind: _StringLit, start: 1, end: 2},
				{typ: _Lparen, lit: "(", bad: false, kind: _StringLit, start: 2, end: 3},
				{typ: _Literal, lit: "1", bad: false, kind: _IntLit, start: 3, end: 4},
				{typ: _Comma, lit: ",", bad: false, kind: _StringLit, start: 4, end: 5},
				{typ: _Literal, lit: "2", bad: false, kind: _IntLit, start: 5, end: 6},
				{typ: _Rparen, lit: ")", bad: false, kind: _StringLit, start: 6, end: 7},
				{typ: _Plain, lit: "aaaa", bad: false, kind: _StringLit, start: 7, end: 11},
				{typ: _EOF, lit: "", bad: false, kind: _StringLit, start: 11, end: 11},
			},
		},
		{
			raw: "#a11aaaa",
			want: []token{
				{typ: _Hash, lit: "#", bad: false, kind: _StringLit, start: 0, end: 1},
				{typ: _Name, lit: "a", bad: false, kind: _StringLit, start: 1, end: 2},
				{typ: _Literal, lit: "11", bad: false, kind: _IntLit, start: 2, end: 4},
				{typ: _Plain, lit: "aaaa", bad: false, kind: _StringLit, start: 4, end: 8},
				{typ: _EOF, lit: "", bad: false, kind: _StringLit, start: 8, end: 8},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.raw, func(t *testing.T) {
			got := make([]token, 0)
			s := newScanner(tc.raw)
			for s.NextToken() {
				// ignore Pos
				s.token.pos = Pos{}
				got = append(got, *s.token)
			}
			if !reflect.DeepEqual(got, tc.want) {
				for _, tk := range got {
					t.Logf("%#v", tk)
				}
				// for i, tk := range got {
				// 	for _, tk := range got {
				// 		t.Logf("%#v", tk)
				// 	}
				// 	var want *token
				// 	if i < len(tc.want) {
				// 		want = &tc.want[i]
				// 	}
				// 	if !reflect.DeepEqual(&tk, want) {
				// 		t.Logf("#%d, want %#v, got %#v", i, want, &tk)
				// 	}
				// }
				t.Fatal("failed")
			}
		})
	}
}
