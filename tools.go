//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/graphql/introspection"
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
