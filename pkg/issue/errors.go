package issue

import "errors"

var ErrNoParent = errors.New("no parent issue has been set")
