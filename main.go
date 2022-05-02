package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	url := flag.String("url", "", "Start Url")
	parallel := flag.Int("parallel", 1, "Number of parallel workers to navigate through site")
	outputFile := flag.String("output-file", "sitemap.xml", "File to write to")
	maxDepth := flag.Int("max-depth", 1, "Max depth of url navigation recursion")
	flag.Parse()

	start := time.Now()

	log.Println(*url)
	log.Println(*parallel)
	log.Println(*outputFile)
	log.Println(*maxDepth)

	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("Duration: %vs\n", duration.Seconds())
}
