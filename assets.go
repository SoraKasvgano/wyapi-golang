package assets

import "embed"

//go:embed frontend/dist/*
var Frontend embed.FS

//go:embed docs/openapi.json
var OpenAPI embed.FS
