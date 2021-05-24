package createrepo

import (
	"fmt"

	"github.com/payfazz/go-errors/v2"
)

var errAccessDenied = fmt.Errorf("access denied")

func IsAccessDenied(err error) bool {
	return errors.Is(err, errAccessDenied)
}
