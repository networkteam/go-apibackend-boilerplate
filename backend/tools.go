//go:build tools
// +build tools

package backend

// Import modules for external tools for correct version pinning und usage with "go run ..."
import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/networkteam/construct/cmd/construct"
	_ "github.com/networkteam/refresh"
	_ "gotest.tools/gotestsum"
)
