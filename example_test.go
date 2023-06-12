package sqls_test

import (
	"fmt"

	"github.com/qjebbs/go-sqls"
)

func Example_select() {
	selects := &sqls.Segment{
		Raw: "SELECT #join('#column', ', ')",
	}
	from := &sqls.Segment{
		Raw: "FROM #t1",
	}
	where := &sqls.Segment{
		Prefix: "WHERE",
		Raw:    "#join('#segment', ' AND ')",
	}
	builder := &sqls.Segment{
		Raw: "#join('#segment', ' ')",
		Segments: []*sqls.Segment{
			selects,
			from,
			where,
		},
	}

	var users sqls.Table = "users"
	selects.WithColumns(users.Expressions("id", "name", "email")...)
	from.WithTables(users)
	where.AppendSegments(&sqls.Segment{
		// (#join('#?', ', ') is also supported
		Raw:     "#c1 IN (#join('#$', ', '))",
		Columns: users.Expressions("id"),
		Args:    []any{1, 2, 3},
	})
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
	update.WithArgs("alice", "alice@example.org")
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
	// [alice alice@example.org 1]
}
