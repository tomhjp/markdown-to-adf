package renderer

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestValidDocument(t *testing.T) {
	b, err := os.ReadFile("testdata/test.md")
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	require.NoError(t, Render(buffer, b))

	schemaContents, err := os.ReadFile("testdata/adf_schema_v1.json")
	require.NoError(t, err)
	schemaLoader := gojsonschema.NewStringLoader(string(schemaContents))
	documentLoader := gojsonschema.NewStringLoader(buffer.String())

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	require.NoError(t, err)

	var errors []string
	for _, desc := range result.Errors() {
		errors = append(errors, desc.String())
	}
	require.NoError(t, os.WriteFile("testdata/actual.json", buffer.Bytes(), 0644))
	require.True(t, result.Valid(), buffer.String(), strings.Join(errors, "\n"))
	fmt.Println(buffer.String())
}
