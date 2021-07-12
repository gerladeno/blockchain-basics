package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
)

const subsidy = 1 / 210000

var ErrInsufficientFunds = errors.New("not enough money")

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

func CreateCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{Txid: []byte{}, Vout: -1, ScriptSig: data}
	txout := TXOutput{Value: subsidy, ScriptPubKey: to}
	tx := Transaction{ID: nil, Vin: []TXInput{txin}, Vout: []TXOutput{txout}}
	if err := tx.SetID(); err != nil {
		return nil, err
	}

	return &tx, nil
}

func CreateUTXOTransaction(from, to string, amount int, bc *Blockchain) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs, err := bc.FindSpendableOutputs(from, amount)
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
			inputs = append(inputs, TXInput{Txid: txID, Vout: out, ScriptSig: from})
		}
	}
	outputs = append(outputs, TXOutput{Value: amount, ScriptPubKey: to})
	if acc > amount {
		outputs = append(outputs, TXOutput{Value: acc - amount, ScriptPubKey: from})
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

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
