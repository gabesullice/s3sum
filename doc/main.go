package main

import (
	"log"
	"os"

	"github.com/gabesullice/s3sum/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	dir := "./man"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	header := &doc.GenManHeader{
		Title:   "S3SUM",
		Section: "1",
	}
	if err := doc.GenManTree(cmd.Root(), header, dir); err != nil {
		log.Fatal(err)
	}
}
