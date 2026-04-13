package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/engine"
	_ "github.com/MadJlzz/maddock/internal/resources/file"
	_ "github.com/MadJlzz/maddock/internal/resources/pkg"
	_ "github.com/MadJlzz/maddock/internal/resources/service"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: maddock-agent apply [--dry-run] <manifest.yaml>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "apply":
		applyCmd := flag.NewFlagSet("apply", flag.ExitOnError)
		dryRun := applyCmd.Bool("dry-run", false, "check-only mode, no changes applied")
		err := applyCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		manifestPath := applyCmd.Arg(0)
		rawManifest, err := os.ReadFile(manifestPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		c, err := catalog.Parse(rawManifest)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ctx := context.Background()
		r := engine.Run(ctx, c, *dryRun)
		fmt.Println(r)
		os.Exit(r.ExitCode())
	default:
		fmt.Println("Usage: maddock-agent apply [--dry-run] <manifest.yaml>")
		os.Exit(1)
	}
}
