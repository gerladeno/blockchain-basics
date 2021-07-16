package blockchain

import (
	"os"
	"strconv"
	"strings"
)

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
		cli.log.Infof("Prev. hash: %x", block.PrevBlockHash)
		cli.log.Infof("Transactions: %v", block.Transactions)
		cli.log.Infof("Hash: %x", block.Hash)
		pow := NewProofOfWork(block)
		cli.log.Infof("PoW: %s", strconv.FormatBool(pow.Validate()))

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

func (cli *CLI) getBalance(address []byte) {
	bc, err := GetBlockchain()
	if err != nil {
		cli.log.Warnf("err getting blockchain: %s", err)
		return
	}
	balance := 0
	pubKeyHash := Base58Encode(address)
	UTXOs, err := bc.FindUTXO(pubKeyHash[1 : len(pubKeyHash)-4])
	if err != nil {
		cli.log.Warnf("err finding unspent transaction: %s", err)
		return
	}
	for _, out := range UTXOs {
		balance += out.Value
	}
	cli.log.Infof("Balance of %s: %d", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc, err := GetBlockchain()
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

func (cli *CLI) createWallet() {
	wallets, err := GetWallets()
	if err != nil {
		cli.log.Warnf("err creating wallets: %s", err)
		return
	}
	address, err := wallets.CreateWallet()
	if err != nil {
		cli.log.Warnf("err creating wallet: %s", err)
		return
	}
	if err = wallets.SaveToFile(); err != nil {
		cli.log.Warnf("err saving to file: %s", err)
	}
	cli.log.Infof("new address created: %s", address)
}

func (cli *CLI) listAddresses() {
	wallets, err := GetWallets()
	if err != nil {
		cli.log.Warnf("err creating wallets: %s", err)
		return
	}
	cli.log.Info(strings.Join(wallets.GetAddresses(), " "))
}
