package templates

import "embed"

// FS contains all embedded template files
// Note: all: is used to include directories starting with _ (like __tests__)
//
//go:embed all:common all:gateway all:auth all:database all:static all:frontend all:k8s
var FS embed.FS
