package main


import (
	"os"

	"github.com/gabesullice/s3sum/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
