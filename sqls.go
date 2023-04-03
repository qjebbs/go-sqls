// Package sqls focuses only on bulding SQL queries by free combination
// of segments. Thus, it works naturally with all sql dialects without
// having to deal with the differences between them. *Segment is the
// only concept you need to learn, unlike any other sql builder or
// ORMs, and it's simple.
package sqls

// Builder is the interface for sql builders.
type Builder interface {
	// Build builds and returns the query and args.
	Build() (query string, args []any, err error)
	// BuildTo builds the query and append args to the argStore.
	BuildTo(argStore *[]any) (query string, err error)
}
