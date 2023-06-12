package sqlb_test

import (
	"github.com/qjebbs/go-sqls/sqlb"
	"github.com/qjebbs/go-sqls/syntax"
	"github.com/qjebbs/go-sqls/util"
)

func Example() {
	q := NewUserQueryBuilder(nil).
		WithIDs([]int64{1, 2, 3})
	q.GetUsers()
}

// Wrap with your own build to provide more friendly APIs.
type UserQueryBuilder struct {
	util.QueryAble
	*sqlb.QueryBuilder
}

var Users = sqlb.NewTable("users", "u")

func NewUserQueryBuilder(db util.QueryAble) *UserQueryBuilder {
	b := sqlb.NewQueryBuilder().
		BindVar(syntax.Dollar).
		Distinct().
		From(Users)
	//  .InnerJoin(...).
	// 	LeftJoin(...).
	// 	LeftJoinOptional(...)
	return &UserQueryBuilder{db, b}
}

func (b *UserQueryBuilder) WithIDs(ids []int64) *UserQueryBuilder {
	b.WhereIn(Users.Column("id"), ids)
	return b
}

func (b *UserQueryBuilder) GetUsers() ([]*User, error) {
	b.Select(Users.Columns("id", "name", "email")...)
	return util.ScanBuilder[*User](b.QueryAble, b.QueryBuilder, func() (*User, []any) {
		r := &User{}
		return r, []interface{}{
			&r.ID, &r.Name, &r.Email,
		}
	})
}

type User struct {
	ID    int64
	Name  string
	Email string
}
