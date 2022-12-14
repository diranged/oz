// Package main is boilerplate top level code for setting up a cobra-based CLI tool.
package ozctl

import (
	"github.com/diranged/oz/internal/cmd/ozctl/cmd"
)

func Main() {
	cmd.Execute()
}
