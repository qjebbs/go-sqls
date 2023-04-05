Package sqls focuses only on bulding SQL queries by free combination
of segments. Thus, it works naturally with all sql dialects without
having to deal with the differences between them. Unlike any other
sql builder or ORMs, Segment is the only concept you need to learn.

# Segment

Segment is the builder for a part of or even the full query, it allows you
to write and combine segments with freedom.

With the help of Segment, we pay attention only to the reference relationships
inside the segment, for example, use "$1" to refer the first element of s.Args.

The syntax of the segment is exactly the same as the syntax of the "database/sql",
plus preprocessing functions support:

	SELECT * FROM foo WHERE id IN ($1, $2, $3) AND #segment(1)
	SELECT * FROM foo WHERE id IN (?, ?, ?) AND #segment(1)
	SELECT * FROM foo WHERE #join('#segment', ' AND ')

# Preprocessing Functions

| name            | description                        | example                    |
| --------------- | ---------------------------------- | -------------------------- |
| c, col, column  | Column by index                    | #c1, #c(1)                 |
| t, table        | Table name / alias by index        | #t1, #t(1)                 |
| s, seg, segment | Segment by index                   | #s1, #s(1)                 |
| join            | Join the template by the separator | #join('#segment', ' AND ') |
| $               | Bindvar, usually used in #join()   | #join('#$', ', ')          |
| ?               | Bindvar, usually used in #join()   | #join('#?', ', ')          |

Note:
  - References in the #join template are functions, not function calls.
  - #c1 is equivalent to #c(1), which is a special syntax to call preprocessing functions when a number is the only argument.

# Examples

See [examples_test.go](./examples_test.go) for more examples.

```go
func Example_select() {
	t := sqls.Table{"users", ""}
	selectFrom := &sqls.Segment{
		Prefix: "",
		// join columns with ', '
		Raw: "SELECT #join('#column', ', ') FROM users",
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

	// select columns
	selectFrom.AppendColumns(t.Columns("id", "name", "email")...)
	// append WHERE condition 1
	where.AppendSegments(&sqls.Segment{
		// (#join('#?', ', ') is also supported
		Raw:     "#c1 IN (#join('#$', ', '))",
		Columns: t.Columns("id"),
		Args:    []any{1, 2, 3},
	})
	// append WHERE condition 2
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
```