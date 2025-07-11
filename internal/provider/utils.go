package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

const MaxWorkers int = 5

func convertSliceToAttrValues[T any](slice []T, converter func(T) attr.Value) []attr.Value {
	values := make([]attr.Value, len(slice))
	for i, item := range slice {
		values[i] = converter(item)
	}
	return values
}

// MergeSlicesWithKeys merges two slices based on a key extraction function for each slice.
// For each item in the primary slice, it checks if an update with the same key exists in the updates slice;
// if so, it uses the converted update, otherwise it keeps the primary item. It then appends any additional
// items from the updates slice (as converted) whose keys were not in the primary slice.
// This preserves the order of primary and adds missing updates at the end.
// Unlike MergeSlices, which relies on Go's equality operator for matching, this function uses user-provided
// key extractor functions, allowing matching and merging of items based on custom keys.
func MergeSlicesWithKeys[T any, U any, K comparable](primary []T, updates []U, primaryKeyExtractor func(T) K, updatesKeyExtractor func(U) K, converter func(U) T) []T {
	updatesMap := make(map[K]U, len(updates))
	for _, item := range updates {
		key := updatesKeyExtractor(item)
		updatesMap[key] = item
	}

	result := make([]T, 0, len(primary)+len(updates))
	processedKeys := make(map[K]bool, len(primary)+len(updates))

	for _, item := range primary {
		key := primaryKeyExtractor(item)
		if updatedItem, exists := updatesMap[key]; exists {
			result = append(result, converter(updatedItem))
		} else {
			result = append(result, item)
		}
		processedKeys[key] = true
	}

	for key, updatedItem := range updatesMap {
		if !processedKeys[key] {
			result = append(result, converter(updatedItem))
		}
	}

	return result
}

// MergeSlices merges two slices based on Go's equality operator.
// Items in updates replace matching items in primary; unmatched updates are appended at the end.
// The order of primary is preserved.
func MergeSlices[T any, U any](primary []T, updates []U, converter func(U) T) []T {
	updatesMap := make(map[interface{}]U, len(updates))
	for _, item := range updates {
		updatesMap[item] = item
	}

	result := make([]T, 0, len(primary)+len(updates))
	processedKeys := make(map[interface{}]bool, len(primary)+len(updates))

	for _, item := range primary {
		if updatedItem, exists := updatesMap[item]; exists {
			result = append(result, converter(updatedItem))
		} else {
			result = append(result, item)
		}
		processedKeys[item] = true
	}

	for key, updatedItem := range updatesMap {
		if !processedKeys[key] {
			result = append(result, converter(updatedItem))
		}
	}

	return result
}
