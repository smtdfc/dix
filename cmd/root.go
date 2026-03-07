package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dix",
	Short: "A zero-magic, compile-time Dependency Injection generator for Go",
	Long: `Dix is a powerful code generation tool 
designed to automate dependency wiring in Go projects.

Guided by the Go philosophy of "Explicit over Implicit," Dix avoids 
runtime reflection and "magic" behavior. Instead, it analyzes your 
source code's AST (Abstract Syntax Tree) to build a dependency graph 
and generates clean, readable Go code to initialize your application.

Key Features:
- Compile-time safety: Catch missing dependencies before you run.
- Zero Reflection: No performance overhead at runtime.
- Transparent: Generated code is standard Go that anyone can read and debug.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
