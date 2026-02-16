package issue

import (
	"github.com/selesy/git-bug-ax/internal/types"
)

// Option is a function that applies an option to an Issue.
type Option func(*Issue) error

// WithPriority creates an option that sets the priority.
func WithPriority(p Priority) Option {
	return func(i *Issue) error {
		i.SetPriority(p)
		return nil
	}
}

// WithStatus creates an option that sets the status.
func WithStatus(s Status) Option {
	return func(i *Issue) error {
		i.SetStatus(s)
		return nil
	}
}

// WithType creates an option that sets the type.
func WithType(t Type) Option {
	return func(i *Issue) error {
		i.SetType(t)
		return nil
	}
}

// WithBlocks creates an option that sets the blocks relationship.
func WithBlocks(blocks types.Set[ID, *ID]) Option {
	return func(i *Issue) error {
		return i.SetBlocks(blocks)
	}
}

// WithReferences creates an option that sets the references relationship.
func WithReferences(references types.Set[ID, *ID]) Option {
	return func(i *Issue) error {
		return i.SetReferences(references)
	}
}

// WithDiscoverer creates an option that sets the discoverer.
func WithDiscoverer(id ID) Option {
	return func(i *Issue) error {
		i.SetDiscoverer(id)
		return nil
	}
}

// WithParent creates an option that sets the parent issue.
func WithParent(id ID) Option {
	return func(i *Issue) error {
		i.SetParent(id)
		return nil
	}
}

// WithResolution creates an option that sets the resolution.
func WithResolution(r Resolution) Option {
	return func(i *Issue) error {
		i.SetResolution(r)
		return nil
	}
}

// WithTitle creates an option that sets the title.
func WithTitle(title string) Option {
	return func(i *Issue) error {
		return i.SetTitle(title)
	}
}

// WithLabels creates an option that sets the labels.
func WithLabels(labels types.Set[types.Label, *types.Label]) Option {
	return func(i *Issue) error {
		return i.SetLabels(labels)
	}
}
