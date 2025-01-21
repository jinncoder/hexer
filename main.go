package main

import (
	"log"

	"github.com/archimoebius/hexer/cli"
)

func main() {

	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
