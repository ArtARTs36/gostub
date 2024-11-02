package main

import (
	"context"
	"fmt"
	"github.com/artarts36/gomodfinder"
	"github.com/artarts36/gostub/internal/cmd"
	"github.com/artarts36/gostub/internal/renderer"
	cli "github.com/artarts36/singlecli"
	"golang.org/x/mod/modfile"
	"path/filepath"
	"strings"
)

const (
	defaultFilename          = "stubs.go"
	defaultFilenamePerMethod = "{{ .Interface.Name.Snake.Value }}_{{ .Method.Name.Snake.Value }}_stub.go"
	defaultFilenamePerType   = "{{ .Interface.Name.Snake.Value }}_stub.go"

	defaultTypeName = "Stub{{ .Interface.Name.Pascal.Value }}"
)

func main() {
	application := &cli.App{
		BuildInfo: &cli.BuildInfo{
			Name: "gostub",
		},
		Args: []*cli.ArgDefinition{
			{
				Name:        "source",
				Required:    true,
				Description: "path to source .go file",
			},
		},
		Opts: []*cli.OptDefinition{
			{
				Name:        "skip-exists",
				Description: "skip exists files",
			},
			{
				Name:        "method-body",
				Description: "method-body: nil-returns, panic",
				WithValue:   true,
			},
			{
				Name:      "package",
				WithValue: true,
			},
			{
				Name:      "filename",
				WithValue: true,
			},
			{
				Name:        "per-method",
				Description: "generate stub file per method",
			},
			{
				Name:      "per-method-filename",
				WithValue: true,
			},
			{
				Name:        "per-type",
				Description: "generate stub file per interface",
			},
			{
				Name:      "per-type-filename",
				WithValue: true,
			},
			{
				Name:      "type-name",
				WithValue: true,
			},
			{
				Name:      "out",
				WithValue: true,
			},
			{
				Name:      "interfaces",
				WithValue: true,
			},
			{
				Name:      "source-go-module",
				WithValue: true,
			},
			{
				Name:      "target-go-module",
				WithValue: true,
			},
		},
		Action: run,
	}

	application.RunWithGlobalArgs(context.Background())
}

func run(ctx *cli.Context) error {
	rend, err := renderer.NewRenderer()
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	command := cmd.NewCommand(rend)

	filename := ctx.Opts["filename"]
	if filename == "" {
		filename = defaultFilename
	}

	perMethodFilename := ctx.Opts["per-method-filename"]
	if perMethodFilename == "" {
		perMethodFilename = defaultFilenamePerMethod
	}

	perTypeFilename := ctx.Opts["per-type-filename"]
	if perTypeFilename == "" {
		perTypeFilename = defaultFilenamePerType
	}

	typeName := ctx.Opts["type-name"]
	if typeName == "" {
		typeName = defaultTypeName
	}

	interfacesString := ctx.Opts["interfaces"]
	interfaces := []string{}
	if interfacesString != "" {
		interfaces = strings.Split(interfacesString, ",")
		for i, s := range interfaces {
			interfaces[i] = strings.Trim(s, " ")
		}
	}

	sourceGoModule := ctx.Opts["source-go-module"]
	if sourceGoModule == "" {
		goMod, err := findGoModule(filepath.Dir(ctx.GetArg("source")))
		if err != nil {
			return err
		}
		sourceGoModule = goMod.Module.Mod.Path
	}

	targetGoModule := ctx.Opts["target-go-module"]
	if targetGoModule == "" {
		goMod, err := findCurrentGoModule()
		if err != nil {
			return err
		}
		targetGoModule = goMod.Module.Mod.Path
	}

	return command.Run(ctx.Context, &cmd.Params{
		Source: ctx.GetArg("source"),

		MethodBody: ctx.Opts["method-body"],
		Package:    ctx.Opts["package"],

		Filename: filename,

		MethodPerFile:     ctx.HasOpt("per-method"),
		PerMethodFilename: perMethodFilename,

		TypePerFile:     ctx.HasOpt("per-type"),
		PerTypeFilename: perTypeFilename,

		TypeName: typeName,

		Out:        ctx.Opts["out"],
		Interfaces: interfaces,
		SkipExists: ctx.HasOpt("skip-exists"),

		SourceGoModule: sourceGoModule,
		TargetGoModule: targetGoModule,
	})
}

func findCurrentGoModule() (*modfile.File, error) {
	return findGoModule("./")
}

func findGoModule(dir string) (*modfile.File, error) {
	goMod, err := gomodfinder.Find(dir, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to find go.mod file: %w", err)
	}
	return goMod, nil
}