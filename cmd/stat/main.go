package main

import (
	"log"
	"os"

	"github.com/remind101/stat"
)

func main() {
	stats, err := stat.New(os.Getenv("DOCKER_HOST"))
	if err != nil {
		log.Fatal(err)
	}

	stats.Adapter = stat.NewL2MetAdapter()

	if err := stats.Run(); err != nil {
		log.Fatal(err)
	}
}
