package metadata

const (
	// Prefix is the prefix prepended to all metadata keys
	Prefix = "gba_"

	// KeyPriority is the metadata key for issue priority
	KeyPriority = Prefix + "priority"
	// KeyStatus is the metadata key for issue status
	KeyStatus = Prefix + "status"
	// KeyType is the metadata key for issue type
	KeyType = Prefix + "type"
	// KeyBlocks is the metadata key for issue blocks relationships
	KeyBlocks = Prefix + "blocks"
	// KeyReferences is the metadata key for issue references relationships
	KeyReferences = Prefix + "references"
	// KeyDiscoverer is the metadata key for issue discoverer identity
	KeyDiscoverer = Prefix + "discoverer"
	// KeyParent is the metadata key for parent issue
	KeyParent = Prefix + "parent"
	// KeyResolution is the metadata key for issue resolution
	KeyResolution = Prefix + "resolution"
)
