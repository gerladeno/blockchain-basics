package blockchain

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CLI struct {
	bc *Blockchain
}

func NewCLI(blockchain *Blockchain) *CLI {
	return &CLI{bc: blockchain}
}

func (cli *CLI) Run() error {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addBlock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	tmp := os.Args[:]
	_ = tmp
	switch strings.ToLower(os.Args[1]) {
	case "addblock":
		if err := addBlockCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
	case "printchain":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
	default:
		cli.printUsage()
	}
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		if err := cli.addBlock(*addBlockData); err != nil {
			return err
		}
	}

	if printChainCmd.Parsed() {
		if err := cli.printChain(); err != nil {
			return err
		}
	}
	return nil
}

func (cli *CLI) addBlock(data string) error {
	return cli.bc.AddBlock(data)
}

func (cli *CLI) printChain() error {
	bci := cli.bc.Iterator()
	for {
		block, err := bci.Next()
		if err != nil {
			return err
		}
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}