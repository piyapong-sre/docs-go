// Package api holds the OpenAPI specification, embedded into the binary
// and served at GET /openapi.yaml.
package api

import _ "embed"

//go:embed openapi.yaml
var OpenAPISpec []byte
