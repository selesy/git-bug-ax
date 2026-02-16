package issue

import (
	"fmt"
	"strings"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/entities/bug"
	"github.com/git-bug/git-bug/entities/identity"
	"github.com/git-bug/git-bug/entity"
	"github.com/git-bug/git-bug/entity/dag"
	"github.com/selesy/git-bug-ax/internal/types"
)

// BugInterface represents a bug entity interface for getting metadata
type BugInterface interface {
	Id() entity.Id
	getSnapshot() *bug.Snapshot
}

// Issue wraps a git-bug bug with additional fields and options.
type Issue struct {
	bugIface  interface{}  // The original bug (either bug.Bug or cache.BugCache)
	bugMeta   BugInterface // Interface for querying metadata
	dirty     bool
	mutations map[string]any
	metadata  map[string]string
}

// bugCacheWrapper wraps cache.BugCache to implement BugInterface
type bugCacheWrapper struct {
	*cache.BugCache
}

func (w *bugCacheWrapper) getSnapshot() *bug.Snapshot {
	return w.Snapshot()
}

// bugBugWrapper wraps bug.Bug to implement BugInterface
type bugBugWrapper struct {
	*bug.Bug
}

func (w *bugBugWrapper) getSnapshot() *bug.Snapshot {
	return w.Compile()
}

// New creates a new Issue wrapper around a bug (can be either bug.Bug or cache.BugCache).
func New(b interface{}) (*Issue, error) {
	if b == nil {
		return nil, fmt.Errorf("bug cannot be nil")
	}

	var bugIf BugInterface

	// Support both cache.BugCache and bug.Bug
	switch v := b.(type) {
	case *cache.BugCache:
		bugIf = &bugCacheWrapper{v}
	case *bug.Bug:
		bugIf = &bugBugWrapper{v}
	default:
		return nil, fmt.Errorf("unsupported bug type: %T", b)
	}

	// Load existing metadata from the bug
	metadata := make(map[string]string)
	snap := bugIf.getSnapshot()
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
		bugIface:  b,
		bugMeta:   bugIf,
		dirty:     false,
		mutations: make(map[string]any),
		metadata:  metadata,
	}, nil
}

func (i *Issue) ID() ID {
	return ID{
		Id: i.bugMeta.Id(),
	}
}

func (i *Issue) Operations() []dag.Operation {
	return i.bugMeta.getSnapshot().Operations
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
	if p, exists := i.metadata["ax_priority"]; exists {
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
	if s, exists := i.metadata["ax_status"]; exists {
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
	if t, exists := i.metadata["ax_type"]; exists {
		var issueType Type
		_ = issueType.UnmarshalText([]byte(t))
		return issueType
	}
	return TypeTask
}

// SetTitle sets the title of the issue using the bug's SetTitle operation.
func (i *Issue) SetTitle(title string) error {
	switch bugVal := i.bugIface.(type) {
	case *cache.BugCache:
		_, err := bugVal.SetTitle(title)
		return err
	case *bug.Bug:
		// For bug.Bug, we need to use the cache.SetTitle function or create the operation directly
		// This assumes there's a helper function in cache or bug package
		return fmt.Errorf("SetTitle not supported for bug.Bug directly; use BugCache instead")
	default:
		return fmt.Errorf("unsupported bug type: %T", i.bugIface)
	}
}

// Title returns the title of the issue from the bug snapshot.
func (i *Issue) Title() string {
	snap := i.bugMeta.getSnapshot()
	return snap.Title
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
	if b, exists := i.metadata["ax_blocks"]; exists {
		result := make(types.Set[ID, *ID])
		if err := result.UnmarshalText([]byte(b)); err != nil {
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
	if d, exists := i.metadata["ax_discoverer"]; exists {
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
	if p, exists := i.metadata["ax_parent"]; exists {
		return NewID(p)
	}
	return ID{}, fmt.Errorf("no parent set")
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
	if r, exists := i.metadata["ax_resolution"]; exists {
		var resolution Resolution
		_ = resolution.UnmarshalText([]byte(r))
		return resolution
	}
	return Resolution{}
}

// Labels returns the current labels of the issue from the bug snapshot.
func (i *Issue) Labels() types.Set[types.Label, *types.Label] {
	snap := i.bugMeta.getSnapshot()
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
	snap := i.bugMeta.getSnapshot()
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
	switch bugVal := i.bugIface.(type) {
	case *cache.BugCache:
		_, _, err := bugVal.ChangeLabels(added, removed)
		return err
	case *bug.Bug:
		return fmt.Errorf("SetLabels not supported for bug.Bug directly; use BugCache instead")
	default:
		return fmt.Errorf("unsupported bug type: %T", i.bugIface)
	}
}

// Commit saves the issue changes to the bug.
func (i *Issue) Commit(identity identity.Interface) error {
	if !i.dirty {
		return nil
	}

	metadata := make(map[string]string)

	// Collect all mutations
	for key, val := range i.mutations {
		if v, ok := val.(string); ok {
			metadata[fmt.Sprintf("ax_%s", key)] = v
		} else if v, ok := val.(fmt.Stringer); ok {
			metadata[fmt.Sprintf("ax_%s", key)] = v.String()
		}
	}

	// Apply metadata to bug based on type
	if len(metadata) > 0 {
		var err error
		switch bugVal := i.bugIface.(type) {
		case *cache.BugCache:
			_, err = bugVal.SetMetadata(bugVal.Id(), metadata)
		case *bug.Bug:
			_, err = bug.SetMetadata(bugVal, identity, 0, bugVal.Id(), metadata)
		default:
			return fmt.Errorf("unsupported bug type in Commit: %T", i.bugIface)
		}

		if err != nil {
			return fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	i.dirty = false
	i.mutations = make(map[string]any)

	return nil
}

// // MarshalJSON implements json.Marshaler.
// func (i *Issue) MarshalJSON() ([]byte, error) {
// 	v := i.View()
// 	return json.Marshal(v)
// }

// // View returns a view struct for display purposes.
// func (i *Issue) View() *IssueView {
// 	return &IssueView{
// 		ID:        i.bugMeta.Id(),
// 		Priority:  i.Priority(),
// 		Status:    i.Status(),
// 		Type:      i.Type(),
// 		Metadata:  i.metadata,
// 		Mutations: i.mutations,
// 	}
// }

// // IssueView represents the public view of an issue
// type IssueView struct {
// 	ID        entity.Id         `json:"id"`
// 	Priority  Priority          `json:"priority"`
// 	Status    Status            `json:"status"`
// 	Type      Type              `json:"type"`
// 	Metadata  map[string]string `json:"metadata"`
// 	Mutations map[string]any    `json:"mutations"`
// }

// // TextCodec is an alias for types.TextCodec
// type TextCodec = types.TextCodec
