package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
)

const subsidy = 10

var ErrInsufficientFunds = errors.New("not enough money")

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func CreateCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{Txid: []byte{}, Vout: -1, PubKey: []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{ID: nil, Vin: []TXInput{txin}, Vout: []TXOutput{*txout}}
	if err := tx.SetID(); err != nil {
		return nil, err
	}

	return &tx, nil
}

func CreateUTXOTransaction(from, to string, amount int, bc *Blockchain) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := GetWallets()
	if err != nil {
		return nil, err
	}
	wallet, err := wallets.GetWallet(from)
	if err != nil {
		return nil, err
	}
	pubKeyHash, err := HashPubKey(wallet.PublicKey)
	if err != nil {
		return nil, err
	}
	acc, validOutputs, err := bc.FindSpendableOutputs(pubKeyHash, amount)
	if err != nil {
		return nil, err
	}
	if acc < amount {
		return nil, ErrInsufficientFunds
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, err
		}
		for _, out := range outs {
			inputs = append(inputs, TXInput{Txid: txID, Vout: out, PubKey: wallet.PublicKey})
		}
	}
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{ID: nil, Vin: inputs, Vout: outputs}
	if err = tx.SetID(); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (tx *Transaction) SetID() error {
	var encoded bytes.Buffer
	var hash [32]byte

	if err := gob.NewEncoder(&encoded).Encode(tx); err != nil {
		return err
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
	return nil
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (in *TXInput) UsesKey(pubKeyHash []byte) (bool, error) {
	lockingHash, err := HashPubKey(in.PubKey)
	if err != nil {
		return false, err
	}
	return bytes.Compare(lockingHash, pubKeyHash) == 0, nil
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Encode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := TXOutput{value, nil}
	txo.Lock([]byte(address))
	return &txo
}
