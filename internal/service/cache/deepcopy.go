package cache

import (
	"encoding/json"
)

func deepCopyStruct[T any](source T, targetPointer *T) error {
	jsonBytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBytes, targetPointer)
	return err
}

func deepCopyStringSlice(original []string) []string {
	// the slice pointers returned by Cacheable already are a snapshot in time because
	// the Cacheable switches the pointer and doesn't ever change existing values

	// now prevent users from inadvertently updating the slice from the outside by making a copy of its values
	result := make([]string, len(original))
	// for strings, which are immutable in go, this _is_ a deep copy
	copy(result, original)
	return result
}
