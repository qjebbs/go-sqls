// Package sqlb is a SQL query builder based on `sqls.Segment`.
package sqlb

import "git.qjebbs.com/jebbs/go-sqls"

// Builder is the interface for sql builders.
type Builder interface {
	// Build builds and returns the query and args.
	Build() (query string, args []any, err error)
	// BuildTo builds the query and append args to the argStore.
	BuildContext(ctx *sqls.Context) (query string, err error)
}
