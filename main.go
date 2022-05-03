package main

import (
	"flag"
	"fmt"
	"github.com/autocorrectoff/SimpleSitemapGenerator/engine"
	"github.com/autocorrectoff/SimpleSitemapGenerator/utils"
	"time"
)

func main() {
	url := flag.String("url", "", "Start Url")
	parallel := flag.Int("parallel", 1, "Number of parallel workers to navigate through site")
	outputFile := flag.String("output-file", "sitemap.xml", "File to write to")
	maxDepth := flag.Int("max-depth", 1, "Max depth of url navigation recursion")
	flag.Parse()

	userInput := engine.UserInput{
		Url:        *url,
		Parallel:   *parallel,
		OutputFile: *outputFile,
		MaxDepth:   *maxDepth,
	}

	start := time.Now()

	siteMap := engine.New(userInput)

	err := siteMap.Start()
	utils.HandleError(err)

	siteMap.Export()

	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("Duration: %vs\n", duration.Seconds())
}
