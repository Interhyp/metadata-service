package cacheable

// Cacheable implements a synchronized in-memory cache for references to arbitrary data structures
//
// Update the timestamp whenever you do a full rescan.
type Cacheable interface {
	GetTimestamp() string
	SetTimestamp(timestamp string)

	GetSortedKeys() *[]string

	// GetEntryRef obtains the reference to the current entry, or nil if the key isn't present.
	GetEntryRef(key string) *interface{}

	// UpdateEntryRef will create the key if it doesn't exist
	// if newRef is nil, will remove the key if present
	//
	// key changes lead to re-creating the sorted keys cache
	UpdateEntryRef(key string, newRef *interface{})
}
