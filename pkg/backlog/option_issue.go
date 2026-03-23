package backlog

import (
	"github.com/selesy/git-bug-agent/internal/codec"
	"github.com/selesy/git-bug-agent/internal/collections"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

type CreateOption interface {
	UpdateOption
	applyCreate(*Issue) error
}

type UpdateOption interface {
	applyUpdate(*Issue) error
}

var _ CreateOption = WithBlocks{}

type WithBlocks collections.Set[issue.ID, *issue.ID]

func (o WithBlocks) applyCreate(i *Issue) error {
	return i.SetBlocks(collections.Set[issue.ID, *issue.ID](o))
}

func (o WithBlocks) applyUpdate(i *Issue) error {
	return i.SetBlocks(collections.Set[issue.ID, *issue.ID](o))
}

var _ UpdateOption = WithDescription{}

type WithDescription issue.Description

func (o WithDescription) applyUpdate(i *Issue) error {
	return i.SetDescription(issue.Description(o))
}

var _ CreateOption = WithDiscoverer{}

type WithDiscoverer issue.ID

func (o WithDiscoverer) applyCreate(i *Issue) error {
	i.SetDiscoverer(issue.ID(o))

	return nil
}

func (o WithDiscoverer) applyUpdate(i *Issue) error {
	i.SetDiscoverer(issue.ID(o))

	return nil
}

var _ CreateOption = WithLabels{}

type WithLabels collections.Set[codec.Label, *codec.Label]

func (o WithLabels) applyCreate(i *Issue) error {
	return i.SetLabels(collections.Set[codec.Label, *codec.Label](o))
}

func (o WithLabels) applyUpdate(i *Issue) error {
	return i.SetLabels(collections.Set[codec.Label, *codec.Label](o))
}

var _ CreateOption = WithParent{}

type WithParent issue.ID

func (o WithParent) applyCreate(i *Issue) error {
	i.SetParent(issue.ID(o))

	return nil
}

func (o WithParent) applyUpdate(i *Issue) error {
	i.SetParent(issue.ID(o))

	return nil
}

var _ CreateOption = WithPriority{}

type WithPriority issue.Priority

func (o WithPriority) applyCreate(i *Issue) error {
	i.SetPriority(issue.Priority(o))

	return nil
}

func (o WithPriority) applyUpdate(i *Issue) error {
	i.SetPriority(issue.Priority(o))

	return nil
}

var _ CreateOption = WithReferences{}

type WithReferences collections.Set[issue.ID, *issue.ID]

func (o WithReferences) applyCreate(i *Issue) error {
	return i.SetReferences(collections.Set[issue.ID, *issue.ID](o))
}

func (o WithReferences) applyUpdate(i *Issue) error {
	return i.SetReferences(collections.Set[issue.ID, *issue.ID](o))
}

var _ UpdateOption = WithResolution{}

type WithResolution issue.Resolution

func (o WithResolution) applyUpdate(i *Issue) error {
	i.SetResolution(issue.Resolution(o))

	return nil
}

var _ UpdateOption = WithStatus{}

type WithStatus issue.Status

func (o WithStatus) applyUpdate(i *Issue) error {
	i.SetStatus(issue.Status(o))

	return nil
}

var _ UpdateOption = WithTitle("")

type WithTitle string

func (o WithTitle) applyUpdate(i *Issue) error {
	return i.SetTitle(string(o))
}

var _ UpdateOption = WithType{}

type WithType issue.Status

func (o WithType) applyUpdate(i *Issue) error {
	i.SetType(issue.Type(o))

	return nil
}
