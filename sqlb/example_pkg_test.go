package sqlb_test

import (
	"github.com/qjebbs/go-sqls"
	"github.com/qjebbs/go-sqls/slices"
	"github.com/qjebbs/go-sqls/sqlb"
	"github.com/qjebbs/go-sqls/syntax"
)

func Example() {
	NewUserQueryBuilder(nil).Search("keyword").GetUsers()
}

// Wrap with your own build to provide more friendly APIs.
type UserQueryBuilder struct {
	*sqlb.QueryBuilder
}

func NewUserQueryBuilder(db sqlb.QueryAble) *UserQueryBuilder {
	b := sqlb.NewQueryBuilder(db).
		BindVar(syntax.Dollar).
		Distinct().
		From(TableUsers)
	// b.InnerJoin( /*...*/ ).
	// 	LeftJoin( /*...*/ ).
	// 	LeftJoinOptional( /*...*/ )
	return &UserQueryBuilder{b}
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
	scanned, err := b.QueryBuilder.Scan(&userScanner{})
	if err != nil {
		return nil, err
	}
	return slices.Atot[*User](scanned), nil
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

func (s *userScanner) NewTarget() (any, []any) {
	r := &User{}
	return r, []interface{}{
		&r.ID, &r.Name, &r.Email,
	}
}
