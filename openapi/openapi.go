package openapi

import _ "embed"

//go:embed openapi.yaml
var OpenapiYAML []byte

//go:embed swagger.html
var SwaggerHTML []byte
