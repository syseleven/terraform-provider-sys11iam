package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func convertSliceToAttrValues[T any](slice []T, converter func(T) attr.Value) []attr.Value {
	values := make([]attr.Value, len(slice))
	for i, item := range slice {
		values[i] = converter(item)
	}
	return values
}
