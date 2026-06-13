package cli

import (
	"errors"

	"github.com/tamnd/comick-cli/comick"
)

func isNotFound(err error) bool {
	return errors.Is(err, comick.ErrNotFound)
}
