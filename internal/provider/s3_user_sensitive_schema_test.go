package provider

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

func schemaForTest(t *testing.T, target fwresource.Resource) fwresource.SchemaResponse {
	t.Helper()

	ctx := context.Background()
	var resp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	return resp
}

func TestProjectS3UserKeyResourceSecretKeyIsSensitive(t *testing.T) {
	resp := schemaForTest(t, &ProjectS3UserKeyResource{})

	secretKey, ok := resp.Schema.Attributes["secret_key"].(resourceschema.StringAttribute)
	require.True(t, ok, "secret_key should be a StringAttribute")
	require.True(t, secretKey.Sensitive, "secret_key should be marked sensitive")
}

func TestProjectS3UserResourceNestedSecretKeyIsSensitive(t *testing.T) {
	resp := schemaForTest(t, &ProjectS3UserResource{})

	keys, ok := resp.Schema.Attributes["keys"].(resourceschema.ListNestedAttribute)
	require.True(t, ok, "keys should be a ListNestedAttribute")

	secretKey, ok := keys.NestedObject.Attributes["secret_key"].(resourceschema.StringAttribute)
	require.True(t, ok, "keys.secret_key should be a StringAttribute")
	require.True(t, secretKey.Sensitive, "keys.secret_key should be marked sensitive")
}
