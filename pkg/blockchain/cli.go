package blockchain

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

type CLI struct {
	bc  *Blockchain
	log *logrus.Logger
}

func NewCLI(log *logrus.Logger, blockchain *Blockchain) *CLI {
	return &CLI{bc: blockchain, log: log}
}

func (cli *CLI) Run() error {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch strings.ToLower(os.Args[1]) {
	case "getbalance":
		if err := getBalanceCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
	case "createblockchain":
		if err := createBlockchainCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
	case "printchain":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
	case "send":
		if err := sendCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
	default:
		cli.printUsage()
		return nil
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			return nil
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			return nil
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			return nil
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
	return nil
}

func (cli *CLI) createBlockchain(address string) {
	bc, err := CreateBlockchain(address)
	if err != nil {
		cli.log.Warnf("err creating blockchain: %s", err)
		return
	}
	if err = bc.db.Close(); err != nil {
		cli.log.Warnf("err closing db: %s", err)
		return
	}
}

func (cli *CLI) addBlock(transactions []*Transaction) error {
	return cli.bc.MineBlock(transactions)
}

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()
	for {
		block, err := bci.Next()
		if err != nil {
			cli.log.Warnf("err getting next block: %s", err)
			return
		}
		cli.log.Infof("Prev. hash: %x\n", block.PrevBlockHash)
		cli.log.Infof("Transactions: %s\n", block.Transactions)
		cli.log.Infof("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		cli.log.Infof("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) printUsage() {
	cli.log.Infof("Usage:")
	cli.log.Infof("  addblock -data BLOCK_DATA - add a block to the blockchain")
	cli.log.Infof("  printchain - print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) getBalance(address string) {
	bc, err := GetBlockchain(address)
	if err != nil {
		cli.log.Warnf("err getting blockchain: %s", err)
		return
	}
	balance := 0
	UTXOs, err := bc.FindUTXO(address)
	if err != nil {
		cli.log.Warnf("err finding unspent transaction: %s", err)
		return
	}
	for _, out := range UTXOs {
		balance += out.Value
	}
	cli.log.Infof("Balance of %s: %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc, err := GetBlockchain(from)
	if err != nil {
		cli.log.Warnf("err getting blockchain: %s", err)
		return
	}
	defer func() {
		err = bc.db.Close()
		if err != nil {
			cli.log.Warnf("err closing db: %s", err)
			return
		}
	}()
	tx, err := CreateUTXOTransaction(from, to, amount, bc)
	if err != nil {
		cli.log.Warnf("err creating transaction: %s", err)
		return
	}
	if err = bc.MineBlock([]*Transaction{tx}); err != nil {
		cli.log.Warnf("err mining block: %s", tx.ID)
		return
	}
}
