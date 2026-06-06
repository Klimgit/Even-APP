package main

import (
	_ "embed"
	"os"
	"text/template"
)

//go:embed openapi_handler.go.tmpl
var tmplText string

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		panic("usage: openapi-handler-gen <out_file> <package_name> <spec_file>")
	}
	outFile := args[0]
	data := struct {
		PackageName string
		SpecFile    string
	}{
		PackageName: args[1],
		SpecFile:    args[2],
	}
	tmpl, err := template.New("openapi_handler").Parse(tmplText)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()
	if err := tmpl.Execute(f, data); err != nil {
		panic(err)
	}
}
