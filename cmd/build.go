package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/smtdfc/dix/generator"
	"github.com/smtdfc/dix/helpers"
	"github.com/smtdfc/dix/parser"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [target] [directory]",
	Short: "Generate wiring code and compile the Go binary",
	Long: `The 'build' command is a shortcut that combines code generation and 
compilation. 


Example:
  dix build main.go ./example
  dix build app.go .`,

	Run: func(cmd *cobra.Command, args []string) {
		targetBuildFile := "main.go"
		targetDir := "."

		if len(args) > 0 {
			targetBuildFile = args[0]
		}

		if len(args) > 1 {
			targetDir = args[1]
		}

		p := parser.NewParser()
		g := generator.NewGenerator()
		mt, err := p.Parse(targetDir)
		if err != nil {
			fatalDixError(err)
		}

		now := time.Now().Unix()
		fileName := fmt.Sprintf("scan_%d.dix", now)
		err = helpers.SaveMetadata(mt, fileName)
		if err != nil {
			fatalDixError(err)
		}

		code, err := g.Generate(mt)
		if err != nil {
			fatalDixError(err)
		}

		err = helpers.WriteTextFile(code, "./generated/dix/root.go")
		if err != nil {
			fatalDixError(err)
		}

		fmt.Println("\033[32m[Build]\033[0m Building ... ")
		command := exec.Command("go", "build", targetBuildFile)

		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fatalDixError(err)
		}

		fmt.Println("\033[32m[Build]\033[0m Build successfully ")

	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

}
