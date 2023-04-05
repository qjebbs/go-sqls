package sqlb_test

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
	"git.qjebbs.com/jebbs/go-sqls/sqlb"
)

func ExampleQueryBuilder_Build() {
	var (
		foo, fooAlias sqls.Table = "foo", "f"
		bar, barAlias sqls.Table = "bar", "b"
	)
	query, args, err := sqlb.NewQueryBuilder(nil).
		Select(fooAlias.Columns("*")).
		From(foo, fooAlias).
		InnerJoin(bar, barAlias, &sqls.Segment{
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				barAlias.Column("foo_id"),
				fooAlias.Column("id"),
			},
		}).
		Where(&sqls.Segment{
			Raw:     "(#c1=$1 OR #c2=$2)",
			Columns: fooAlias.Columns("a", "b"),
			Args:    []any{1, 2},
		}).
		Where2(barAlias.Column("c"), "=", 3).
		Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT f.* FROM foo AS f INNER JOIN bar AS b ON b.foo_id=f.id WHERE (f.a=$1 OR f.b=$2) AND b.c=$3
	// [1 2 3]
}

func ExampleQueryBuilder_LeftJoinOptional() {
	var (
		foo, fooAlias sqls.Table = "foo", "f"
		bar, barAlias sqls.Table = "bar", "b"
	)
	query, args, err := sqlb.NewQueryBuilder(nil).
		Distinct(). // *QueryBuilder trims optional joins only when SELECT DISTINCT is used.
		Select(fooAlias.Columns("*")).
		From(foo, fooAlias).
		// declare an optional LEFT JOIN
		LeftJoinOptional(bar, barAlias, &sqls.Segment{
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				barAlias.Column("foo_id"),
				fooAlias.Column("id"),
			},
		}).
		// don't touch any columns of "bar", so that it can be trimmed
		Where2(fooAlias.Column("id"), ">", 1).
		Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT DISTINCT f.* FROM foo AS f WHERE f.id>$1
	// [1]
}

func ExampleQueryBuilder_With() {
	var (
		foo, fooAlias sqls.Table = "foo", "f"
		bar, barAlias sqls.Table = "bar", "b"
		cte, cteAlias sqls.Table = "bar_type_1", "b1"
	)
	query, args, err := sqlb.NewQueryBuilder(nil).
		With(cte, cteAlias, &sqls.Segment{
			Raw:     "SELECT * FROM #t1 AS #t2 WHERE #c1=$1",
			Columns: barAlias.Columns("type"),
			Args:    []any{1},
			Tables:  []sqls.Table{bar, barAlias},
		}).
		Select([]*sqls.TableColumn{
			fooAlias.Column("*"),
			cteAlias.Column("*"),
		}).
		From(foo, fooAlias).
		LeftJoin(cte, cteAlias, &sqls.Segment{
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				cteAlias.Column("foo_id"),
				fooAlias.Column("id"),
			},
		}).
		Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// With bar_type_1 AS (SELECT * FROM bar AS b WHERE b.type=$1) SELECT f.*, b1.* FROM foo AS f LEFT JOIN bar_type_1 AS b1 ON b1.foo_id=f.id
	// [1]
}

func ExampleQueryBuilder_Union() {
	var foo, fooAlias sqls.Table = "foo", "f"
	columns := fooAlias.Columns("*")
	query, args, err := sqlb.NewQueryBuilder(nil).
		Select(columns).
		From(foo, fooAlias).
		Where2(fooAlias.Column("id"), " = ", 1).
		Union(
			sqlb.NewQueryBuilder(nil).
				From(foo, fooAlias).
				WhereIn(fooAlias.Column("id"), []any{2, 3, 4}).
				Select(columns),
		).
		Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT f.* FROM foo AS f WHERE f.id = $1 UNION (SELECT f.* FROM foo AS f WHERE f.id IN ($2, $3, $4))
	// [1 2 3 4]
}
