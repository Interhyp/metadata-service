package cache

import (
	openapi "github.com/Interhyp/metadata-service/api/v1"
)

// we could do this generically using reflection, but it's much slower and harder to understand

func deepCopyOwner(immutableOwnerPtrNonNil *interface{}) openapi.OwnerDto {
	// the pointers returned by Cacheable already are a snapshot in time because
	// the Cacheable switches the pointer and doesn't ever change existing values

	// now prevent users from updating the struct in the cache by making a deep copy of its values
	firstLevelCopy := (*immutableOwnerPtrNonNil).(openapi.OwnerDto)
	// Note: strings are immutable in Go, so we don't need to duplicate *string either

	return firstLevelCopy
}

func deepCopyService(immutableServicePtrNonNil *interface{}) openapi.ServiceDto {
	// the pointers returned by Cacheable already are a snapshot in time because
	// the Cacheable switches the pointer and doesn't ever change existing values

	// now prevent users from updating the struct in the cache by making a deep copy of its values
	firstLevelCopy := (*immutableServicePtrNonNil).(openapi.ServiceDto)
	// Note: strings are immutable in Go, so we don't need to duplicate *string either

	firstLevelCopy.Quicklinks = deepCopyQuicklinkSlice(firstLevelCopy.Quicklinks)
	firstLevelCopy.Repositories = deepCopyStringSlice(firstLevelCopy.Repositories)
	firstLevelCopy.DevelopmentOnly = deepCopyBoolPtr(firstLevelCopy.DevelopmentOnly)

	return firstLevelCopy
}

func deepCopyRepository(immutableRepositoryPtrNonNil *interface{}) openapi.RepositoryDto {
	// the pointers returned by Cacheable already are a snapshot in time because
	// the Cacheable switches the pointer and doesn't ever change existing values

	// now prevent users from updating the struct in the cache by making a deep copy of its values
	firstLevelCopy := (*immutableRepositoryPtrNonNil).(openapi.RepositoryDto)
	// Note: strings are immutable in Go, so we don't need to duplicate *string either

	firstLevelCopy.Configuration = deepCopyRepositoryConfiguration(firstLevelCopy.Configuration)

	return firstLevelCopy
}

// --- substructures ---

func deepCopyQuicklink(original openapi.Quicklink) openapi.Quicklink {
	// strings are immutable in go, so can just copy all the string pointers here
	return original
}

func deepCopyQuicklinkSlice(original []openapi.Quicklink) []openapi.Quicklink {
	result := make([]openapi.Quicklink, len(original))
	for i, v := range original {
		result[i] = deepCopyQuicklink(v)
	}
	return result
}

func deepCopyRepositoryConfiguration(original *openapi.RepositoryConfigurationDto) *openapi.RepositoryConfigurationDto {
	if original == nil {
		return nil
	}

	firstLevelCopy := *original

	if firstLevelCopy.Approvers != nil {
		approversCopy := make(map[string][]string)
		for key, value := range *firstLevelCopy.Approvers {
			approversCopy[key] = deepCopyStringSlice(value)
		}
		firstLevelCopy.Approvers = &approversCopy
	}

	return &firstLevelCopy
}

// --- helpers ---

func deepCopyStringSlice(original []string) []string {
	// the slice pointers returned by Cacheable already are a snapshot in time because
	// the Cacheable switches the pointer and doesn't ever change existing values

	// now prevent users from inadvertently updating the slice from the outside by making a copy of its values
	result := make([]string, len(original))
	// for strings, which are immutable in go, this _is_ a deep copy
	copy(result, original)
	return result
}

func deepCopyBoolPtr(original *bool) *bool {
	if original == nil {
		return original
	}
	value := *original
	return &value
}
