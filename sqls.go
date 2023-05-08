// Package sqls focuses only on bulding SQL queries by free combination
// of segments. Thus, it works naturally with all sql dialects without
// having to deal with the differences between them. Unlike any other
// sql builder or ORMs, Segment is the only concept you need to learn.
//
// # Segment
//
// Segment is the builder for a part of or even the full query, it allows you
// to write and combine segments with freedom.
//
// With the help of Segment, we pay attention only to the reference relationships
// inside the segment, for example, use "$1" to refer the first element of s.Args.
//
// The syntax of the segment is exactly the same as the syntax of the "database/sql",
// plus preprocessing functions support:
//
//	SELECT * FROM foo WHERE id IN ($1, $2, $3) AND #segment(1)
//	SELECT * FROM foo WHERE id IN (?, ?, ?) AND #segment(1)
//	SELECT * FROM foo WHERE #join('#segment', ' AND ')
//
// # Preprocessing Functions
//
//   - c, col, column 		: Column by index, e.g. #c1, #c(1)
//   - t, table				: Table name / alias by index, e.g. #t1, #t(1)
//   - s, seg, segment		: Segment by index, e.g. #s1, #s(1)
//   - join 				: Join the template by the separator, e.g. #join('#column', ', '), #join('#c=#$', ', ')
//   - $ 					: Argument by index, used in #join().
//   - ?					: Argument by index, used in #join().
//
// Note:
//   - References in the #join template are functions, not function calls.
//   - #c1 is equivalent to #c(1), which is a special syntax to call preprocessing functions when a number is the only argument.
package sqls

// Builder is the interface for sql builders.
type Builder interface {
	// Build builds and returns the query and args.
	Build() (query string, args []any, err error)
	// BuildTo builds the query and append args to the argStore.
	BuildContext(ctx *Context) (query string, err error)
}
