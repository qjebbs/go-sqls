package sqls_test

import (
	"fmt"

	"github.com/qjebbs/go-sqls"
)

func Example_select() {
	selectFrom := &sqls.Segment{
		Prefix: "",
		// join columns with ', '
		Raw: "SELECT #join('#column', ', ') FROM #t1",
	}
	where := &sqls.Segment{
		Prefix: "WHERE",
		// join segments with ' AND '
		Raw: "#join('#segment', ' AND ')",
	}
	builder := &sqls.Segment{
		// join segments with ' '
		Raw: "#join('#segment', ' ')",
		Segments: []*sqls.Segment{
			selectFrom,
			where,
		},
	}

	var users sqls.Table = "users"
	selectFrom.WithTables(users)
	// select columns
	selectFrom.AppendColumns(users.Expressions("id", "name", "email")...)
	// append WHERE condition 1
	where.AppendSegments(&sqls.Segment{
		// (#join('#?', ', ') is also supported
		Raw:     "#c1 IN (#join('#$', ', '))",
		Columns: users.Expressions("id"),
		Args:    []any{1, 2, 3},
	})
	// append WHERE condition 2
	where.AppendSegments(&sqls.Segment{
		Raw:     "#c1 = $1",
		Columns: users.Expressions("active"),
		Args:    []any{true},
	})

	bulit, args, err := builder.Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(bulit)
	fmt.Println(args)
	// Output:
	// SELECT id, name, email FROM users WHERE id IN ($1, $2, $3) AND active = $4
	// [1 2 3 true]
}

func Example_update() {
	update := &sqls.Segment{
		Prefix: "",
		Raw:    "UPDATE #t1 SET #join('#c=#$', ', ')",
	}
	where := &sqls.Segment{
		Prefix: "WHERE",
		Raw:    "#join('#segment', ' AND ')",
	}
	// consider wrapping it with your own builder
	// to provide a more friendly APIs
	builder := &sqls.Segment{
		Raw: "#join('#segment', ' ')",
		Segments: []*sqls.Segment{
			update,
			where,
		},
	}

	var users sqls.Table = "users"
	update.WithTables(users)
	update.WithColumns(users.Expressions("name", "email")...)
	update.WithArgs("jebbs", "qjebbs@gmail.com")
	// append as many conditions as you want
	where.AppendSegments(&sqls.Segment{
		Raw:     "#c1=$1",
		Columns: users.Expressions("id"),
		Args:    []any{1},
	})

	bulit, args, err := builder.Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(bulit)
	fmt.Println(args)
	// Output:
	// UPDATE users SET name=$1, email=$2 WHERE id=$3
	// [jebbs qjebbs@gmail.com 1]
}

func ExampleInterpolate() {
	query := "SELECT * FROM foo WHERE id IN ($1, $2, $3) AND status=$4"
	args := []any{1, 2, 3, "ok"}
	interpolated, err := sqls.Interpolate(query, args...)
	if err != nil {
		panic(err)
	}
	fmt.Println(interpolated)
	// Output:
	// SELECT * FROM foo WHERE id IN (1, 2, 3) AND status='ok'
}
