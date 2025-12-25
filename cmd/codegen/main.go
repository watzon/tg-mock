// cmd/codegen/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	specPath := flag.String("spec", "spec/api.json", "path to API spec")
	outDir := flag.String("out", "gen", "output directory")
	flag.Parse()

	spec, err := loadSpec(*specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded spec: %s (%s)\n", spec.Version, spec.ReleaseDate)
	fmt.Printf("Methods: %d, Types: %d\n", len(spec.Methods), len(spec.Types))
	fmt.Printf("Output directory: %s\n", *outDir)

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	if err := generateTypes(spec, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate types: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated types.go")

	if err := generateMethods(spec, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate methods: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated methods.go")

	if err := generateFixtures(spec, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate fixtures: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated fixtures.go")
}

func loadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}
