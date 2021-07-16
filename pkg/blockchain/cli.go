package blockchain

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
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
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
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
	case "createwallet":
		if err := createWalletCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
	case "listaddresses":
		if err := listAddressesCmd.Parse(os.Args[2:]); err != nil {
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
		cli.getBalance([]byte(*getBalanceAddress))
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			return nil
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
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
