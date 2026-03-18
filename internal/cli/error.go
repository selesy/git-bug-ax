package cli

import "errors"

// ErrWrongPrefixCount is returned when a command receives more than one issue prefix argument.
var ErrWrongPrefixCount = errors.New("only a single issue prefix is expected")
