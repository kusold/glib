package errx_test

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/kusold/glib/errx"
)

func ExampleWrap() {
	err := errx.Wrap(sql.ErrNoRows, errx.CodeNotFound, "user not found")

	code, _ := errx.CodeOf(err)
	fmt.Println(code)
	fmt.Println(errx.PublicMessage(err))
	fmt.Println(errors.Is(err, sql.ErrNoRows))

	// Output:
	// not_found
	// user not found
	// true
}
