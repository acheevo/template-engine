package main

import (
	"github.com/acheevo/template-engine/cmd"
	_ "github.com/acheevo/template-engine/internal/templates" // Register template types
)

func main() {
	cmd.Execute()
}
