package v1

import _ "embed"

//go:embed api.yaml
var OpenAPISpec []byte // skeleton OpenAPI document
