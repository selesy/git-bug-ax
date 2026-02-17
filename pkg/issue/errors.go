package issue

import "errors"

var ErrNoParent = errors.New("no parent issue has been set")

var ErrNoTitle = errors.New("an WithTitle option is required during issue creation")
