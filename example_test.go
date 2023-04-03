package sqls_test

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
)

func ExampleSegment_select() {
	t := sqls.Table{"users", ""}
	selectFrom := &sqls.Segment{
		Header: "",
		Raw:    "SELECT #join('#column', ', ') FROM users",
	}
	where := &sqls.Segment{
		Header: "WHERE",
		Raw:    "#join('#segment', ' AND ')",
	}
	builder := &sqls.Segment{
		Raw: "#join('#segment', ' ')",
		Segments: []*sqls.Segment{
			selectFrom,
			where,
		},
	}

	selectFrom.AppendColumns(t.Columns("id", "name", "email")...)
	// append as many conditions as you want
	where.AppendSegments(&sqls.Segment{
		Raw:     "#c1 IN (#join('#$', ', '))", // (#join('#?', ', ') is also supported
		Columns: t.Columns("id"),
		Args:    []any{1, 2, 3},
	})
	where.AppendSegments(&sqls.Segment{
		Raw:     "#c1 = $1",
		Columns: t.Columns("active"),
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

func ExampleSegment_update() {
	t := sqls.Table{"users", ""}
	update := &sqls.Segment{
		Header: "",
		Raw:    "UPDATE users SET #join('#c=#$', ', ')",
	}
	where := &sqls.Segment{
		Header: "WHERE",
		Raw:    "#join('#segment', ' AND ')",
	}
	builder := &sqls.Segment{
		Raw: "#join('#segment', ' ')",
		Segments: []*sqls.Segment{
			update,
			where,
		},
	}

	update.AppendColumns(t.Columns("name", "email")...)
	update.AppendArgs("jebbs", "qjebbs@gmail.com")
	// append as many conditions as you want
	where.AppendSegments(&sqls.Segment{
		Raw:     "#c1=$1",
		Columns: t.Columns("id"),
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
