package sqlb_test

import (
	"github.com/qjebbs/go-sqls"
	"github.com/qjebbs/go-sqls/sqlb"
	"github.com/qjebbs/go-sqls/syntax"
	"github.com/qjebbs/go-sqls/util"
)

func Example() {
	NewUserQueryBuilder(nil).Search("keyword").GetUsers()
}

// Wrap with your own build to provide more friendly APIs.
type UserQueryBuilder struct {
	util.QueryAble
	*sqlb.QueryBuilder
}

func NewUserQueryBuilder(db util.QueryAble) *UserQueryBuilder {
	b := sqlb.NewQueryBuilder().
		BindVar(syntax.Dollar).
		Distinct().
		From(TableUsers)
	// b.InnerJoin( /*...*/ ).
	// 	LeftJoin( /*...*/ ).
	// 	LeftJoinOptional( /*...*/ )
	return &UserQueryBuilder{db, b}
}

func (b *UserQueryBuilder) Search(keyword string) *UserQueryBuilder {
	b.Where(&sqls.Segment{
		// read the document in `sqls` package to learn how to write a segment.
		Raw:     "(#c1 like '%' || $1 || '%' OR #c2 like '%' || $1 || '%')",
		Columns: TableUsers.Columns("name", "email"),
		Args:    []any{keyword},
	})
	return b
}

func (b *UserQueryBuilder) GetUsers() ([]*User, error) {
	scanner := &userScanner{}
	b.Select(scanner.Select()...)
	return util.ScanBuilder[*User](b.QueryAble, b.QueryBuilder, scanner.NewScanTarget)
}

var TableUsers = sqlb.NewTable("users", "u")

type User struct {
	ID    int64
	Name  string
	Email string
}

type userScanner struct{}

func (s *userScanner) Select() []*sqls.TableColumn {
	return TableUsers.Columns("id", "name", "email")
}

func (s *userScanner) NewScanTarget() (*User, []any) {
	r := &User{}
	return r, []interface{}{
		&r.ID, &r.Name, &r.Email,
	}
}
