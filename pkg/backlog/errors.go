package backlog

import "errors"

var ErrNoCreate = errors.New("expected the first operation to be create")
