package sqlb_test

import (
	"reflect"
	"testing"

	"git.qjebbs.com/jebbs/go-sqls"
	"git.qjebbs.com/jebbs/go-sqls/sqlb"
)

func TestQueryBuilder(t *testing.T) {
	var (
		users, usersAlias sqls.Table = "users", "u"
		foo, fooAlias     sqls.Table = "foo", "f"
		bar, barAlias     sqls.Table = "bar", "b"
	)
	q := sqlb.NewQueryBuilder(nil).Distinct().
		With(users, usersAlias, &sqls.Segment{
			Raw:  "SELECT * FROM users WHERE type=$1",
			Args: []any{"user"},
		}).
		With("xxx", "x", &sqls.Segment{Raw: "SELECT 1 AS whatever"}) // should be ignored
	q.Select(fooAlias.Columns("id", "name")).
		From(users, usersAlias).
		LeftJoinOptional(foo, fooAlias, &sqls.Segment{
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				fooAlias.Column("user_id"),
				usersAlias.Column("id"),
			},
		}).
		LeftJoinOptional(bar, barAlias, &sqls.Segment{ // not referenced, should be ignored
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				barAlias.Column("user_id"),
				usersAlias.Column("id"),
			},
		}).
		Where2(usersAlias.Column("id"), "=", 1).
		Union(
			sqlb.NewQueryBuilder(nil).
				Select(fooAlias.Columns("id", "name")).
				From(foo, fooAlias).
				Where(&sqls.Segment{
					Raw:     "#c1>$1 AND #c1<$2",
					Columns: fooAlias.Columns("id"),
					Args:    []any{10, 20},
				}),
		)
	gotQuery, gotArgs, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	wantQuery := "With users AS (SELECT * FROM users WHERE type=$1) SELECT DISTINCT f.id, f.name FROM users AS u LEFT JOIN foo AS f ON f.user_id=u.id WHERE u.id=$2 UNION (SELECT f.id, f.name FROM foo AS f WHERE f.id>$3 AND f.id<$4)"
	wantArgs := []any{"user", 1, 10, 20}
	if wantQuery != gotQuery {
		t.Errorf("want:\n%s\ngot:\n%s", wantQuery, gotQuery)
	}
	if !reflect.DeepEqual(wantArgs, gotArgs) {
		t.Errorf("want:\n%v\ngot:\n%v", wantArgs, gotArgs)
	}
}
