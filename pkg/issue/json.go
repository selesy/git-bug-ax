package issue

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/git-bug/git-bug/entities/bug"
	"github.com/git-bug/git-bug/entities/common"
	"github.com/git-bug/git-bug/entity"
	"github.com/git-bug/git-bug/entity/dag"

	"github.com/selesy/git-bug-ax/internal/codec"
	"github.com/selesy/git-bug-ax/internal/collections"
)

func (i *Issue) MarshalJSON() ([]byte, error) {
	view, err := newView(i)
	if err != nil {
		return nil, err
	}

	ops := i.Operations()
	if len(ops) == 0 {
		return nil, ErrNoCreate
	}

	firstOp := ops[0]
	if _, ok := firstOp.(*bug.CreateOperation); !ok {
		return nil, ErrNoCreate
	}

	created := firstOp.Time()
	view.Created = &created

	for _, op := range ops {
		view.History = append(view.History, newHistory(op))
	}

	updated := ops[len(ops)-1].Time()
	view.Updated = &updated

	return json.Marshal(view)
}

func (i *Issue) Excerpt() *Excerpt {
	return &Excerpt{iss: i}
}

var _ json.Marshaler = (*Excerpt)(nil)

type Excerpt struct {
	iss *Issue
}

func (e *Excerpt) MarshalJSON() ([]byte, error) {
	view, err := newView(e.iss)
	if err != nil {
		return nil, err
	}

	return json.Marshal(view)
}

// view is the lightweight format returned by list operations (ready, blocked, mine, etc).
// It excludes body/description and computed fields, focusing on metadata needed for filtering and coordination.
type view struct {
	ID          ID                                         `json:"id"`
	Title       string                                     `json:"title"`
	Description Description                                `json:"description"`
	Type        Type                                       `json:"type"`
	Status      Status                                     `json:"status"`
	Priority    Priority                                   `json:"priority"`
	Assignee    interface{}                                `json:"assignee"` // issue.Assignee or null
	Discoverer  ID                                         `json:"discoverer,omitempty"`
	Parent      ID                                         `json:"parent,omitempty"`
	Blocks      collections.Set[ID, *ID]                   `json:"blocks,omitempty"`
	Labels      collections.Set[codec.Label, *codec.Label] `json:"labels"`
	Created     *time.Time                                 `json:"created,omitempty"`
	Updated     *time.Time                                 `json:"updated,omitempty"`
	History     []history                                  `json:"history,omitempty"`
}

func newView(i *Issue) (*view, error) {
	blocks, err := i.Blocks()
	if err != nil {
		return nil, err
	}

	// Not having a Discoverer is not an error in this context
	discoverer, err := i.Discoverer()
	if errors.Is(err, ErrNoDiscoverer) {
		err = nil
	}

	if err != nil {
		return nil, err
	}

	// Not having a parent is not an error in this context
	parent, err := i.Parent()
	if errors.Is(err, ErrNoParent) {
		err = nil
	}

	if err != nil {
		return nil, err
	}

	return &view{
		ID:          i.ID(),
		Title:       i.Title(),
		Description: i.Description(),
		Type:        i.Type(),
		Status:      i.Status(),
		Priority:    i.Priority(),
		// TODO: Assignee
		Discoverer: discoverer,
		Parent:     parent,
		Blocks:     blocks,
		Labels:     i.Labels(),
	}, nil
}

type history struct {
	ID     entity.Id
	Type   dag.OperationType `json:"type"`
	Name   string            `json:"name"`
	Time   string            `json:"time"`
	Change map[string]any    `json:"change,omitempty"`
	Author entity.Id         `json:"author"`
}

func newHistory(op dag.Operation) history {
	var (
		name   string
		change map[string]any
	)

	switch op := op.(type) {
	case *bug.CreateOperation:
		name = "CreateOperation"
		change = map[string]any{
			"title":       op.Title,
			"description": op.Message,
			// TODO: show filenames?
		}
	case *bug.SetTitleOperation:
		name = "SetTitleOperation"
		change = map[string]any{
			"title": op.Title,
		}
	case *bug.AddCommentOperation:
		name = "AddCommentOperation"
		change = map[string]any{
			"message": op.Message,
			// TODO: show filenames?
		}
	case *bug.SetStatusOperation:
		name = "SetStatusOperation"
		change = map[string]any{
			"status": op.Status.String(),
		}
	case *bug.LabelChangeOperation:
		name = "LabelChangeOperation"
		change = map[string]any{
			"added":   labels(op.Added),
			"removed": labels(op.Removed),
		}
	case *bug.EditCommentOperation:
		name = "Changed comment"
		change = map[string]any{
			"message": op.Message,
			// TODO: show filenames?
		}
	case *dag.NoOpOperation[*bug.Snapshot]:
		name = "NoOpOperation"
	case *dag.SetMetadataOperation[*bug.Snapshot]:
		name = "SetMetadataOperation"
		change = make(map[string]any, len(op.NewMetadata))
		for k, v := range op.NewMetadata {
			change[k] = any(v)
		}
	}

	return history{
		ID:     op.Id(),
		Type:   op.Type(),
		Name:   name,
		Time:   op.Time().UTC().Format(time.RFC3339),
		Change: change,
		Author: op.Author().Id(),
	}
}

func labels(labels []common.Label) []string {
	var s []string
	for _, l := range labels {
		s = append(s, l.String())

	}

	return s
}
