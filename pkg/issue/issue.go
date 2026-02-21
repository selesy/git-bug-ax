package issue

import (
	"errors"
	"fmt"
	"strings"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/commands/execenv"
	"github.com/git-bug/git-bug/entities/bug"
	"github.com/git-bug/git-bug/entities/identity"
	"github.com/git-bug/git-bug/entity/dag"
	"github.com/selesy/git-bug-ax/internal/metadata"
	"github.com/selesy/git-bug-ax/internal/types"
)

const defaultDescription = `
# {title}

{Overview paragraph describing the purpose and context of this work.}

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

// Issue wraps a git-bug bug with additional fields and options.
type Issue struct {
	bug       *cache.BugCache
	dirty     bool
	mutations map[string]any
	metadata  map[string]string
}

func Create(env *execenv.Env, opts ...Option) (*Issue, error) {
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
		return nil, ErrNoTitle
	}

	// TODO: should we add a default description if one is not provided?
	_ = defaultDescription

	bug, _, err := env.Backend.Bugs().New(issWrap.createTitle, issWrap.createDescription.description)
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

func (i *Issue) Update(opts ...Option) (*Issue, error) {
	wrapper := newIssueWrapper(i)

	var errs error
	for _, opt := range opts {
		errs = errors.Join(errs, opt.fn(wrapper))
	}

	return i, errs
}

func (i *Issue) ID() ID {
	return ID{
		Id: i.bug.Id(),
	}
}

func (i *Issue) Operations() []dag.Operation {
	return i.bug.Snapshot().Operations
}

// IsDirty returns true if the issue has been modified.
func (i *Issue) IsDirty() bool {
	return i.dirty
}

// SetPriority sets the priority of the issue.
func (i *Issue) SetPriority(p Priority) {
	i.mutations["priority"] = p
	i.dirty = true
}

// Priority returns the current priority of the issue.
func (i *Issue) Priority() Priority {
	if p, ok := i.mutations["priority"]; ok {
		if priority, ok := p.(Priority); ok {
			return priority
		}
	}
	// Try to get from metadata
	if p, exists := i.metadata[metadata.KeyPriority]; exists {
		var priority Priority
		_ = priority.UnmarshalText([]byte(p))
		return priority
	}
	return PriorityMedium
}

// SetStatus sets the status of the issue.
func (i *Issue) SetStatus(s Status) {
	i.mutations["status"] = s
	i.dirty = true
}

// Status returns the current status of the issue.
func (i *Issue) Status() Status {
	if s, ok := i.mutations["status"]; ok {
		if status, ok := s.(Status); ok {
			return status
		}
	}
	if s, exists := i.metadata[metadata.KeyStatus]; exists {
		var status Status
		_ = status.UnmarshalText([]byte(s))
		return status
	}
	return StatusDraft
}

// SetType sets the type of the issue.
func (i *Issue) SetType(t Type) {
	i.mutations["type"] = t
	i.dirty = true
}

// Type returns the current type of the issue.
func (i *Issue) Type() Type {
	if t, ok := i.mutations["type"]; ok {
		if typ, ok := t.(Type); ok {
			return typ
		}
	}
	if t, exists := i.metadata[metadata.KeyType]; exists {
		var issueType Type
		_ = issueType.UnmarshalText([]byte(t))
		return issueType
	}
	return TypeTask
}

// SetTitle sets the title of the issue using the bug's SetTitle operation.
func (i *Issue) SetTitle(title string) error {
	_, err := i.bug.SetTitle(title)
	return err
}

// Title returns the title of the issue from the bug snapshot.
func (i *Issue) Title() string {
	snap := i.bug.Snapshot()
	return snap.Title
}

// Description returns the "body" of the issue
func (i *Issue) Description() Description {
	comments := i.bug.Snapshot().Comments
	if len(comments) < 1 {
		return Description{description: ""}
	}

	return Description{description: comments[0].Message}
}

// SetDescription sets the issue's description (first comment)
func (i *Issue) SetDescription(description Description) error {
	_, _, err := i.bug.EditCreateComment(description.description)

	return err
}

// SetBlocks sets the blocks relationship.
func (i *Issue) SetBlocks(blocks types.Set[ID, *ID]) error {
	data, err := blocks.MarshalText()
	if err != nil {
		return err
	}
	i.mutations["blocks"] = string(data)
	i.dirty = true
	return nil
}

// Blocks returns the current blocks relationship.
func (i *Issue) Blocks() (types.Set[ID, *ID], error) {
	if b, ok := i.mutations["blocks"]; ok {
		if blocksStr, ok := b.(string); ok {
			result := make(types.Set[ID, *ID])
			if err := result.UnmarshalText([]byte(blocksStr)); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	if b, exists := i.metadata[metadata.KeyBlocks]; exists {
		result := make(types.Set[ID, *ID])
		if err := result.UnmarshalText([]byte(b)); err != nil {
			return nil, err
		}
		return result, nil
	}
	return make(types.Set[ID, *ID]), nil
}

// SetReferences sets the references relationship.
func (i *Issue) SetReferences(references types.Set[ID, *ID]) error {
	data, err := references.MarshalText()
	if err != nil {
		return err
	}
	i.mutations["references"] = string(data)
	i.dirty = true
	return nil
}

// References returns the current references relationship.
func (i *Issue) References() (types.Set[ID, *ID], error) {
	if r, ok := i.mutations["references"]; ok {
		if referencesStr, ok := r.(string); ok {
			result := make(types.Set[ID, *ID])
			if err := result.UnmarshalText([]byte(referencesStr)); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	if r, exists := i.metadata[metadata.KeyReferences]; exists {
		result := make(types.Set[ID, *ID])
		if err := result.UnmarshalText([]byte(r)); err != nil {
			return nil, err
		}
		return result, nil
	}
	return make(types.Set[ID, *ID]), nil
}

// SetDiscoverer sets the discoverer identity.
func (i *Issue) SetDiscoverer(id ID) {
	i.mutations["discoverer"] = id.String()
	i.dirty = true
}

// Discoverer returns the current discoverer identity.
func (i *Issue) Discoverer() (ID, error) {
	if d, ok := i.mutations["discoverer"]; ok {
		if discovererStr, ok := d.(string); ok {
			return NewID(discovererStr)
		}
	}
	if d, exists := i.metadata[metadata.KeyDiscoverer]; exists {
		return NewID(d)
	}
	return ID{}, fmt.Errorf("no discoverer set")
}

// SetParent sets the parent issue.
func (i *Issue) SetParent(id ID) {
	i.mutations["parent"] = id.String()
	i.dirty = true
}

// Parent returns the current parent issue.
func (i *Issue) Parent() (ID, error) {
	if p, ok := i.mutations["parent"]; ok {
		if parentStr, ok := p.(string); ok {
			return NewID(parentStr)
		}
	}
	if p, exists := i.metadata[metadata.KeyParent]; exists {
		return NewID(p)
	}
	return ID{}, ErrNoParent
}

// SetResolution sets the resolution of the issue.
func (i *Issue) SetResolution(r Resolution) {
	i.mutations["resolution"] = r
	i.dirty = true
}

// Resolution returns the current resolution of the issue.
func (i *Issue) Resolution() Resolution {
	if r, ok := i.mutations["resolution"]; ok {
		if resolution, ok := r.(Resolution); ok {
			return resolution
		}
	}
	if r, exists := i.metadata[metadata.KeyResolution]; exists {
		var resolution Resolution
		_ = resolution.UnmarshalText([]byte(r))
		return resolution
	}
	return Resolution{}
}

// Labels returns the current labels of the issue from the bug snapshot.
func (i *Issue) Labels() types.Set[types.Label, *types.Label] {
	snap := i.bug.Snapshot()
	result := make(types.Set[types.Label, *types.Label])
	for _, label := range snap.Labels {
		result.Add(types.Label(label.String()))
	}
	return result
}

// SetLabels sets the labels of the issue using the bug's ChangeLabels operation.
// It determines which labels have been added or removed (case-insensitive) and
// creates a LabelChangeOperation if changes are detected.
func (i *Issue) SetLabels(newLabels types.Set[types.Label, *types.Label]) error {
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
