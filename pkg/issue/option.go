package issue

import (
	"github.com/selesy/git-bug-ax/internal/types"
)

type issueWrapper struct {
	i                 *Issue
	createTitle       string
	createDescription Description
}

func newIssueWrapper(iss *Issue) *issueWrapper {
	return &issueWrapper{
		i:                 iss,
		createTitle:       "",
		createDescription: Description{description: ""},
	}
}

// Option is a function that applies an option to an Issue.
type Option struct {
	fn    func(*issueWrapper) error
	newFN func(*issueWrapper) error
}

// WithBlocks creates an option that sets the blocks relationship.
func WithBlocks(blocks types.Set[ID, *ID]) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			return i.i.SetBlocks(blocks)
		},
	}
}

// WithTitle creates an option that sets the title.
func WithDescription(description Description) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			return i.i.SetDescription(description)
		},
		newFN: func(i *issueWrapper) error {
			i.createDescription = description

			return nil
		},
	}
}

// WithDiscoverer creates an option that sets the discoverer.
func WithDiscoverer(id ID) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			i.i.SetDiscoverer(id)
			return nil
		},
	}
}

// WithLabels creates an option that sets the labels.
func WithLabels(labels types.Set[types.Label, *types.Label]) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			return i.i.SetLabels(labels)
		},
	}
}

// WithParent creates an option that sets the parent issue.
func WithParent(id ID) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			i.i.SetParent(id)
			return nil
		},
	}
}

// WithPriority creates an option that sets the priority.
func WithPriority(p Priority) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			i.i.SetPriority(p)
			return nil
		},
	}
}

// WithReferences creates an option that sets the references relationship.
func WithReferences(references types.Set[ID, *ID]) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			return i.i.SetReferences(references)
		},
	}
}

// WithResolution creates an option that sets the resolution.
func WithResolution(r Resolution) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			i.i.SetResolution(r)
			return nil
		},
	}
}

// WithStatus creates an option that sets the status.
func WithStatus(s Status) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			i.i.SetStatus(s)
			return nil
		},
	}
}

// WithTitle creates an option that sets the title.
func WithTitle(title string) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			i.createTitle = title
			return i.i.SetTitle(title)
		},
		newFN: func(i *issueWrapper) error {
			if title == "" {
				return ErrNoTitle
			}

			i.createTitle = title

			return nil
		},
	}
}

// WithType creates an option that sets the type.
func WithType(t Type) Option {
	return Option{
		fn: func(i *issueWrapper) error {
			i.i.SetType(t)
			return nil
		},
	}
}
