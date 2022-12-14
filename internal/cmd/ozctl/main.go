// Package ozctl is the top level entrypoint for the cobra-based CLI tool
package ozctl

import (
	"github.com/diranged/oz/internal/cmd/ozctl/cmd"
)

// Main begins the command execution
func Main() {
	cmd.Execute()
}
