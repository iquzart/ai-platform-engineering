package docs

import _ "embed"

// OpenAPI contains the API contract served by the Swagger UI.
//
//go:embed openapi.yaml
var OpenAPI []byte
