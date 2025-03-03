package main

import (
	"flag"
	"fmt"
	"github.com/csotherden/go-tmpl-ssg/pkg/generator"
	"log"
	"os"
)

func main() {
	templateDir := flag.String("template", "", "Path to the template directory")
	outputDir := flag.String("output", "", "Path to the output directory")
	flag.Parse()

	if *templateDir == "" || *outputDir == "" {
		log.Fatal("Both --template and --output flags must be provided")
	}

	if err := os.MkdirAll(*outputDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	err := generator.GenerateSite(*templateDir, *outputDir)
	if err != nil {
		log.Fatalf("Error generating site: %v", err)
	}

	fmt.Println("Site generation completed successfully!")
}
