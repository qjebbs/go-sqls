package sqlb_test

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
	"git.qjebbs.com/jebbs/go-sqls/sqlb"
	"git.qjebbs.com/jebbs/go-sqls/syntax"
)

func ExampleQueryBuilder_Build() {
	var (
		foo = sqlb.NewTable("foo", "f")
		bar = sqlb.NewTable("bar", "b")
	)
	b := sqlb.NewQueryBuilder(nil).
		Select(foo.Columns("*")).
		From(foo).
		InnerJoin(bar, &sqls.Segment{
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				bar.Column("foo_id"),
				foo.Column("id"),
			},
		}).
		Where(&sqls.Segment{
			Raw:     "(#c1=$1 OR #c2=$1)",
			Columns: foo.Columns("a", "b"),
			Args:    []any{1},
		}).
		Where2(bar.Column("c"), "=", 2)

	query, args, err := b.BindVar(syntax.Dollar).Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
	fmt.Println(args)
	query, args, err = b.BindVar(syntax.Question).Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT f.* FROM foo AS f INNER JOIN bar AS b ON b.foo_id=f.id WHERE (f.a=$1 OR f.b=$1) AND b.c=$2
	// [1 2]
	// SELECT f.* FROM foo AS f INNER JOIN bar AS b ON b.foo_id=f.id WHERE (f.a=? OR f.b=?) AND b.c=?
	// [1 1 2]
}

func ExampleQueryBuilder_LeftJoinOptional() {
	var (
		foo = sqlb.NewTable("foo", "f")
		bar = sqlb.NewTable("bar", "b")
	)
	query, args, err := sqlb.NewQueryBuilder(nil).
		BindVar(syntax.Dollar).
		Distinct(). // *QueryBuilder trims optional joins only when SELECT DISTINCT is used.
		Select(foo.Columns("*")).
		From(foo).
		// declare an optional LEFT JOIN
		LeftJoinOptional(bar, &sqls.Segment{
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				bar.Column("foo_id"),
				foo.Column("id"),
			},
		}).
		// don't touch any columns of "bar", so that it can be trimmed
		Where2(foo.Column("id"), ">", 1).
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
		foo = sqlb.NewTable("foo", "f")
		bar = sqlb.NewTable("bar", "b")
		cte = sqlb.NewTable("bar_type_1", "b1")
	)
	query, args, err := sqlb.NewQueryBuilder(nil).
		BindVar(syntax.Dollar).
		With(cte.Name, &sqls.Segment{
			Raw:     "SELECT * FROM #t1 AS #t2 WHERE #c1=$1",
			Columns: bar.Columns("type"),
			Args:    []any{1},
			Tables:  bar.Names(),
		}).
		Select([]*sqls.TableColumn{
			foo.Column("*"),
			cte.Column("*"),
		}).
		From(foo).
		LeftJoinOptional(cte, &sqls.Segment{
			Raw: "#c1=#c2",
			Columns: []*sqls.TableColumn{
				cte.Column("foo_id"),
				foo.Column("id"),
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
	var foo = sqlb.NewTable("foo", "f")
	columns := foo.Columns("*")
	query, args, err := sqlb.NewQueryBuilder(nil).
		BindVar(syntax.Dollar).
		Select(columns).
		From(foo).
		Where2(foo.Column("id"), " = ", 1).
		Union(
			sqlb.NewQueryBuilder(nil).
				BindVar(syntax.Dollar).
				From(foo).
				WhereIn(foo.Column("id"), []any{2, 3, 4}).
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
