package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

var ErrNoSuchWallet = errors.New("err no such wallet")

type Wallets struct {
	Wallets map[string]*Wallet
}

func GetWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	if err := wallets.LoadFromFile(); err != nil {
		return nil, err
	}
	return &wallets, nil
}

func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := CreateWallet()
	if err != nil {
		return "", err
	}
	address, err := wallet.GetAddress()
	if err != nil {
		return "", err
	}
	ws.Wallets[string(address)] = wallet
	return string(address), nil
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws *Wallets) GetWallet(address string) (*Wallet, error) {
	wallet, ok := ws.Wallets[address]
	if !ok {
		return nil, fmt.Errorf("%w %s", ErrNoSuchWallet, address)
	}
	return wallet, nil
}

func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return ws.SaveToFile()
	}
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}
	var wallets Wallets
	gob.Register(elliptic.P256())
	err = gob.NewDecoder(bytes.NewReader(fileContent)).Decode(&wallets)
	if err != nil {
		return err
	}
	ws.Wallets = wallets.Wallets
	return err
}

func (ws Wallets) SaveToFile() error {
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	if err := gob.NewEncoder(&content).Encode(ws); err != nil {
		return err
	}
	if err := ioutil.WriteFile(walletFile, content.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}
