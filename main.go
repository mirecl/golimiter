package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
	"github.com/mirecl/golimiter/linters"
)

// Version golimiter linter.
const Version string = "0.8.1"

func main() {
	jsonFlag := flag.Bool("json", false, "format report")
	versionFlag := flag.Bool("version", false, "version golimiter")
	configFlag := flag.String("config", ".golimiter.yaml", "path config file")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("golimiter %s\n", Version)
		return
	}

	cfg, err := config.ReadFromFile(*configFlag)
	if err != nil {
		panic(err)
	}

	allIssues := analysis.Run(cfg, linters.All...)

	if *jsonFlag {
		if allIssuesBytes, err := json.Marshal(allIssues); err == nil {
			fmt.Println(string(allIssuesBytes))
		}
		return
	}

	for linter, issues := range allIssues {
		for _, issue := range issues {
			position := fmt.Sprintf("%s:%v", analysis.GetPathRelative(issue.Filename), issue.Line)
			fmt.Printf("%s \033[31m%s: %s. \033[0m\033[30m(%s)\033[0m\n", position, linter, issue.Message, issue.Hash)
		}
	}
}
