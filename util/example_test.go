package util_test

import (
	"fmt"
	"time"

	"github.com/qjebbs/go-sqls/util"
)

func ExampleInterpolate() {
	query := "SELECT * FROM foo WHERE status = ? AND created_at > ?"
	args := []any{"ok", time.Unix(0, 0)}
	interpolated, err := util.Interpolate(query, args, util.WithTimeFormat("2006-01-02 15:04:05"))
	if err != nil {
		panic(err)
	}
	fmt.Println(interpolated)
	// Output:
	// SELECT * FROM foo WHERE status = 'ok' AND created_at > '1970-01-01 08:00:00'
}
