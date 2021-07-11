package main

import (
	"blockchain-basics/pkg/blockchain"
)

func main() {
	bc, err := blockchain.NewBlockchain()
	if err != nil {
		panic(err)
	}
	cli := blockchain.NewCLI(bc)
	if err := cli.Run(); err != nil {
		panic(err)
	}
}
