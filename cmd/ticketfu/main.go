package main

import (
	"os"

	"github.com/taonic/ticketfu/cli"
)

func main() {
	if err := cli.Run(os.Args); err != nil {
		panic(err)
	}
}
