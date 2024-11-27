package main

import (
	"os"

	"github.com/dqx0/blockchain/cmd"
)

func main() {
	defer os.Exit(0)

	cli := cmd.CommandLine{}
	cli.Run()
}
