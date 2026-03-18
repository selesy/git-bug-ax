package backlog

import (
	"github.com/selesy/git-bug-agent/internal/codec"
	"github.com/selesy/git-bug-agent/internal/collections"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

type issueWrapper struct {
	i                 *Issue
	createTitle       string
	createDescription issue.Description
}

func newIssueWrapper(iss *Issue) *issueWrapper {
	return &issueWrapper{
		i:                 iss,
		createTitle:       "",
		createDescription: issue.Description{},
	}
}

// IssueOption is a function that applies an option to an Issue.
type IssueOption struct {
	fn    func(*issueWrapper) error
	newFN func(*issueWrapper) error
}

// WithBlocks creates an option that sets the blocks relationship.
func WithBlocks(blocks collections.Set[issue.ID, *issue.ID]) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			return i.i.SetBlocks(blocks)
		},
	}
}

// WithTitle creates an option that sets the title.
func WithDescription(description issue.Description) IssueOption {
	return IssueOption{
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
func WithDiscoverer(id issue.ID) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			i.i.SetDiscoverer(id)
			return nil
		},
	}
}

// WithLabels creates an option that sets the labels.
func WithLabels(labels collections.Set[codec.Label, *codec.Label]) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			return i.i.SetLabels(labels)
		},
	}
}

// WithParent creates an option that sets the parent issue.
func WithParent(id issue.ID) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			i.i.SetParent(id)
			return nil
		},
	}
}

// WithPriority creates an option that sets the priority.
func WithPriority(p issue.Priority) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			i.i.SetPriority(p)
			return nil
		},
	}
}

// WithReferences creates an option that sets the references relationship.
func WithReferences(references collections.Set[issue.ID, *issue.ID]) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			return i.i.SetReferences(references)
		},
	}
}

// WithResolution creates an option that sets the resolution.
func WithResolution(r issue.Resolution) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			i.i.SetResolution(r)
			return nil
		},
	}
}

// WithStatus creates an option that sets the status.
func WithStatus(s issue.Status) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			i.i.SetStatus(s)
			return nil
		},
	}
}

// WithTitle creates an option that sets the title.
func WithTitle(title string) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			i.createTitle = title
			return i.i.SetTitle(title)
		},
		newFN: func(i *issueWrapper) error {
			if title == "" {
				return issue.ErrNoTitle
			}

			i.createTitle = title

			return nil
		},
	}
}

// WithType creates an option that sets the type.
func WithType(t issue.Type) IssueOption {
	return IssueOption{
		fn: func(i *issueWrapper) error {
			i.i.SetType(t)
			return nil
		},
	}
}
