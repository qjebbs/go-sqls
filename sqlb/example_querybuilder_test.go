package sqlb_test

import (
	"git.qjebbs.com/jebbs/go-sqls"
	"git.qjebbs.com/jebbs/go-sqls/slices"
	"git.qjebbs.com/jebbs/go-sqls/sqlb"
)

func ExampleQueryBuilder() {
	_, _ = NewUserQueryBuilder(nil).
		WhereMatches("keyword").
		GetUsers()
}

var (
	TableUsers      sqls.Table = "users"
	TableUsersAlias sqls.Table = "u"
)

type User struct {
	ID    int64
	Name  string
	Email string
}

// It's recommended to wrap it with your struct to provide a more
// friendly API and improve segment reusability.
type UserQueryBuilder struct {
	*sqlb.QueryBuilder
}

// Quick creatation of *UserQueryBuilder and reuse Join segments.
func NewUserQueryBuilder(db sqlb.QueryAble) *UserQueryBuilder {
	b := sqlb.NewQueryBuilder(db).
		Distinct().
		From(TableUsers, TableUsersAlias)
	// b.InnerJoin( /*...*/ ).
	// 	LeftJoin( /*...*/ ).
	// 	LeftJoin( /*...*/ )
	return &UserQueryBuilder{b}
}

// The extending methods provide a more friendly API and improve segments reusability.
func (b *UserQueryBuilder) WhereMatches(keyword string) *UserQueryBuilder {
	b.Where(&sqls.Segment{
		Raw:     "(#c1 like '%' || $1 || '%' OR #c2 like '%' || $1 || '%')",
		Columns: TableUsersAlias.Columns("name", "email"),
		Args:    []any{keyword},
	})
	return b
}

// friendly API to scan and returns []*User instead of []any
func (b *UserQueryBuilder) GetUsers() ([]*User, error) {
	scanned, err := b.QueryBuilder.Scan(&userScanner{})
	if err != nil {
		return nil, err
	}
	return slices.Atot[*User](scanned), nil
}

// userScanner implements sqls.Scanner
type userScanner struct{}

// Select tells *QueryBuilder which columns to select
func (s *userScanner) Select() []*sqls.TableColumn {
	return TableUsersAlias.Columns("id", "name", "email")
}

// NewTarget create a new *User for scanning
func (s *userScanner) NewTarget() (any, []any) {
	r := &User{}
	return r, []interface{}{
		&r.ID, &r.Name, &r.Email,
	}
}
