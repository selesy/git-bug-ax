package issue

import "errors"

var ErrNoDiscoverer = errors.New("no discoverer has been set")

var ErrNoParent = errors.New("no parent issue has been set")

var ErrNoTitle = errors.New("an WithTitle option is required during issue creation")
