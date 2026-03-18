package backlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/commands/execenv"
	"github.com/git-bug/git-bug/entities/bug"
	"github.com/git-bug/git-bug/entities/identity"
	"github.com/git-bug/git-bug/entity/dag"

	"github.com/selesy/git-bug-agent/internal/codec"
	"github.com/selesy/git-bug-agent/internal/collections"
	"github.com/selesy/git-bug-agent/internal/metadata"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

const defaultDescription = `{Overview paragraph describing the purpose and context of this work.}

## Scope

- Brief description of what this task accomplishes
- Boundaries of the work

## Files Affected

- pkg/api/handler.go
- pkg/validate/rules.go

## Environment

- Details for reproducing the development environment
- Required dependencies, environment variables, or secrets
- Link to a devcontainer definition or Dockerfile

## Implementation Notes

- Use existing validator package
- Add new validation rules for email format
- Update handler to call validator before processing

## Acceptance Criteria

- [ ] All user inputs validated before processing
- [ ] Invalid inputs return 400 with descriptive error
- [ ] Unit tests cover new validation rules
- [ ] Existing tests pass

## Verification

- ` + "`go test ./pkg/api -run TestUserValidation`" + `
`

var (
	_ json.Marshaler = (*Issue)(nil)
	// _ json.Unmarshaler = (*Issue)(nil)
)

// Issue wraps a git-bug bug with additional fields and options.
type Issue struct {
	bug       *cache.BugCache
	dirty     bool
	mutations map[string]any
	metadata  map[string]string
}

// Create creates a new Issue with the given options.
func Create(env *execenv.Env, opts ...IssueOption) (*Issue, error) {
	issWrap := newIssueWrapper(&Issue{})

	var errs error
	for _, opt := range opts {
		if opt.newFN != nil {
			errs = errors.Join(errs, opt.newFN(issWrap))
		}
	}

	if errs != nil {
		return nil, errs
	}

	if issWrap.createTitle == "" {
		return nil, issue.ErrNoTitle
	}

	// TODO: should we add a default description if one is not provided?
	_ = defaultDescription

	bug, _, err := env.Backend.Bugs().New(issWrap.createTitle, issWrap.createDescription.Raw())
	if err != nil {
		return nil, err
	}

	iss, err := Wrap(bug)
	if err != nil {
		return nil, err
	}

	// TODO: can this be reused?
	issWrap = newIssueWrapper(iss)

	// var errs error
	for _, opt := range opts {
		if opt.newFN != nil {
			continue
		}

		errs = errors.Join(errs, opt.fn(issWrap))
	}

	if errs != nil {
		return nil, errs
	}

	return iss, nil
}

// Wrap creates a new Issue wrapper around a BugCache.
func Wrap(b *cache.BugCache) (*Issue, error) {
	if b == nil {
		return nil, fmt.Errorf("bug cannot be nil")
	}

	// Load existing metadata from the bug
	metadata := make(map[string]string)
	snap := b.Snapshot()
	for _, op := range snap.Operations {
		metaOp, ok := op.(*dag.SetMetadataOperation[*bug.Snapshot])
		if !ok {
			continue
		}

		for k, v := range metaOp.NewMetadata {
			metadata[k] = v
		}
	}

	return &Issue{
		bug:       b,
		dirty:     false,
		mutations: make(map[string]any),
		metadata:  metadata,
	}, nil
}

// Blocks returns the current blocks relationship.
func (i *Issue) Blocks() (collections.Set[issue.ID, *issue.ID], error) {
	if b, ok := i.mutations["blocks"]; ok {
		if blocksStr, ok := b.(string); ok {
			result := make(collections.Set[issue.ID, *issue.ID])
			if err := result.UnmarshalText([]byte(blocksStr)); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	if b, exists := i.metadata[metadata.KeyBlocks]; exists {
		result := make(collections.Set[issue.ID, *issue.ID])
		if err := result.UnmarshalText([]byte(b)); err != nil {
			return nil, err
		}
		return result, nil
	}
	return make(collections.Set[issue.ID, *issue.ID]), nil
}

// SetBlocks sets the blocks relationship.
func (i *Issue) SetBlocks(blocks collections.Set[issue.ID, *issue.ID]) error {
	data, err := blocks.MarshalText()
	if err != nil {
		return err
	}
	i.mutations["blocks"] = string(data)
	i.dirty = true
	return nil
}

// Description returns the "body" of the issue
func (i *Issue) Description() issue.Description {
	comments := i.bug.Snapshot().Comments
	desc := issue.Description{}
	// TODO: handle err
	_ = desc.UnmarshalText([]byte(comments[0].Message))

	return desc
}

// SetDescription sets the issue's description (first comment)
func (i *Issue) SetDescription(description issue.Description) error {
	_, _, err := i.bug.EditCreateComment(description.Raw())

	return err
}

// Discoverer returns the current discoverer identity.
func (i *Issue) Discoverer() (issue.ID, error) {
	if d, ok := i.mutations["discoverer"]; ok {
		if discovererStr, ok := d.(string); ok {
			return issue.NewID(discovererStr)
		}
	}
	if d, exists := i.metadata[metadata.KeyDiscoverer]; exists {
		return issue.NewID(d)
	}
	return issue.ID{}, issue.ErrNoDiscoverer
}

// SetDiscoverer sets the discoverer identity.
func (i *Issue) SetDiscoverer(id issue.ID) {
	i.mutations["discoverer"] = id.String()
	i.dirty = true
}

// ID returns the ID of the issue.
func (i *Issue) ID() issue.ID {
	return issue.ID{
		Id: i.bug.Id(),
	}
}

// Labels returns the current labels of the issue from the bug snapshot.
func (i *Issue) Labels() collections.Set[codec.Label, *codec.Label] {
	snap := i.bug.Snapshot()
	result := make(collections.Set[codec.Label, *codec.Label])
	for _, label := range snap.Labels {
		result.Add(codec.Label(label.String()))
	}
	return result
}

// SetLabels sets the labels of the issue using the bug's ChangeLabels operation.
// It determines which labels have been added or removed (case-insensitive) and
// creates a LabelChangeOperation if changes are detected.
func (i *Issue) SetLabels(newLabels collections.Set[codec.Label, *codec.Label]) error {
	// Get current labels from the bug snapshot and convert to case-insensitive map
	snap := i.bug.Snapshot()
	currentLabelsLower := make(map[string]struct{})
	for _, label := range snap.Labels {
		currentLabelsLower[strings.ToLower(label.String())] = struct{}{}
	}

	// Convert new labels to case-insensitive map and determine added/removed
	newLabelsLower := make(map[string]struct{})
	var added []string
	for label := range newLabels {
		lower := strings.ToLower(string(label))
		newLabelsLower[lower] = struct{}{}
		if _, exists := currentLabelsLower[lower]; !exists {
			// This is a new label (case-insensitive)
			added = append(added, string(label))
		}
	}

	// Find removed labels
	var removed []string
	for label := range currentLabelsLower {
		if _, exists := newLabelsLower[label]; !exists {
			// Find the original case of the removed label
			for _, origLabel := range snap.Labels {
				if strings.ToLower(origLabel.String()) == label {
					removed = append(removed, origLabel.String())
					break
				}
			}
		}
	}

	// If no changes, return early
	if len(added) == 0 && len(removed) == 0 {
		return nil
	}

	// Apply the label changes to the bug
	_, _, err := i.bug.ChangeLabels(added, removed)
	return err
}

// Parent returns the current parent issue.
func (i *Issue) Parent() (issue.ID, error) {
	if p, ok := i.mutations["parent"]; ok {
		if parentStr, ok := p.(string); ok {
			return issue.NewID(parentStr)
		}
	}
	if p, exists := i.metadata[metadata.KeyParent]; exists {
		return issue.NewID(p)
	}
	return issue.ID{}, issue.ErrNoParent
}

// SetParent sets the parent issue.
func (i *Issue) SetParent(id issue.ID) {
	i.mutations["parent"] = id.String()
	i.dirty = true
}

// Priority returns the current priority of the issue.
func (i *Issue) Priority() issue.Priority {
	if p, ok := i.mutations["priority"]; ok {
		if priority, ok := p.(issue.Priority); ok {
			return priority
		}
	}
	// Try to get from metadata
	if p, exists := i.metadata[metadata.KeyPriority]; exists {
		var priority issue.Priority
		_ = priority.UnmarshalText([]byte(p))
		return priority
	}
	return issue.PriorityMedium
}

// SetPriority sets the priority of the issue.
func (i *Issue) SetPriority(p issue.Priority) {
	i.mutations["priority"] = p
	i.dirty = true
}

// References returns the current references relationship.
func (i *Issue) References() (collections.Set[issue.ID, *issue.ID], error) {
	if r, ok := i.mutations["references"]; ok {
		if referencesStr, ok := r.(string); ok {
			result := make(collections.Set[issue.ID, *issue.ID])
			if err := result.UnmarshalText([]byte(referencesStr)); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	if r, exists := i.metadata[metadata.KeyReferences]; exists {
		result := make(collections.Set[issue.ID, *issue.ID])
		if err := result.UnmarshalText([]byte(r)); err != nil {
			return nil, err
		}
		return result, nil
	}
	return make(collections.Set[issue.ID, *issue.ID]), nil
}

// SetReferences sets the references relationship.
func (i *Issue) SetReferences(references collections.Set[issue.ID, *issue.ID]) error {
	data, err := references.MarshalText()
	if err != nil {
		return err
	}
	i.mutations["references"] = string(data)
	i.dirty = true
	return nil
}

// Resolution returns the current resolution of the issue.
func (i *Issue) Resolution() issue.Resolution {
	if r, ok := i.mutations["resolution"]; ok {
		if resolution, ok := r.(issue.Resolution); ok {
			return resolution
		}
	}
	if r, exists := i.metadata[metadata.KeyResolution]; exists {
		var resolution issue.Resolution
		_ = resolution.UnmarshalText([]byte(r))
		return resolution
	}
	return issue.Resolution{}
}

// SetResolution sets the resolution of the issue.
func (i *Issue) SetResolution(r issue.Resolution) {
	i.mutations["resolution"] = r
	i.dirty = true
}

// Status returns the current status of the issue.
func (i *Issue) Status() issue.Status {
	if s, ok := i.mutations["status"]; ok {
		if status, ok := s.(issue.Status); ok {
			return status
		}
	}
	if s, exists := i.metadata[metadata.KeyStatus]; exists {
		var status issue.Status
		_ = status.UnmarshalText([]byte(s))
		return status
	}
	return issue.StatusDraft
}

// SetStatus sets the status of the issue.
func (i *Issue) SetStatus(s issue.Status) {
	i.mutations["status"] = s
	i.dirty = true
}

// Title returns the title of the issue from the bug snapshot.
func (i *Issue) Title() string {
	snap := i.bug.Snapshot()
	return snap.Title
}

// SetTitle sets the title of the issue using the bug's SetTitle operation.
func (i *Issue) SetTitle(title string) error {
	_, err := i.bug.SetTitle(title)
	return err
}

// Type returns the current type of the issue.
func (i *Issue) Type() issue.Type {
	if t, ok := i.mutations["type"]; ok {
		if typ, ok := t.(issue.Type); ok {
			return typ
		}
	}
	if t, exists := i.metadata[metadata.KeyType]; exists {
		var issueType issue.Type
		_ = issueType.UnmarshalText([]byte(t))
		return issueType
	}
	return issue.TypeTask
}

// SetType sets the type of the issue.
func (i *Issue) SetType(t issue.Type) {
	i.mutations["type"] = t
	i.dirty = true
}

// Commit saves the issue changes to the bug.
func (i *Issue) Commit(identity identity.Interface) error {
	if !i.dirty {
		return nil
	}

	metadataMap := make(map[string]string)

	// Collect all mutations
	for key, val := range i.mutations {
		if v, ok := val.(string); ok {
			metadataMap[metadata.Prefix+key] = v
		} else if v, ok := val.(fmt.Stringer); ok {
			metadataMap[metadata.Prefix+key] = v.String()
		}
	}

	// Apply metadata to bug
	if len(metadataMap) != 0 {
		_, err := i.bug.SetMetadata(i.bug.Id(), metadataMap)
		if err != nil {
			return fmt.Errorf("failed to set metadata: %w", err)
		}
		err = i.bug.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	}

	i.dirty = false
	i.mutations = make(map[string]any)

	return nil
}

// IsDirty returns true if the issue has been modified.
func (i *Issue) IsDirty() bool {
	return i.dirty
}

// Operations returns the operations for the issue.
func (i *Issue) Operations() []dag.Operation {
	return i.bug.Snapshot().Operations
}

// Update updates the issue with the given options.
func (i *Issue) Update(opts ...IssueOption) (*Issue, error) {
	wrapper := newIssueWrapper(i)

	var errs error
	for _, opt := range opts {
		errs = errors.Join(errs, opt.fn(wrapper))
	}

	return i, errs
}
