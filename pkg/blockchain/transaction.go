package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const subsidy = 10

var ErrInsufficientFunds = errors.New("err not enough money")
var ErrIncorrectTransaction = errors.New("err incorrect transaction")
var ErrTransactionNotFound = errors.New("err transaction not found")

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
	var err error

	txin := TXInput{Txid: []byte{}, Vout: -1, PubKey: []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{ID: nil, Vin: []TXInput{txin}, Vout: []TXOutput{*txout}}
	if tx.ID, err = tx.Hash(); err != nil {
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
	if tx.ID, err = tx.Hash(); err != nil {
		return nil, err
	}
	if err = bc.SignTransaction(&tx, wallet.PrivateKey); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (in *TXInput) UsesKey(pubKeyHash []byte) (bool, error) {
	lockingHash, err := HashPubKey(in.PubKey)
	if err != nil {
		return false, err
	}
	return bytes.Equal(lockingHash, pubKeyHash), nil
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := TXOutput{value, nil}
	txo.Lock([]byte(address))
	return &txo
}

func (tx *Transaction) Sing(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}
	var err error
	for _, vin := range tx.Vin {
		if val, ok := prevTXs[hex.EncodeToString(vin.Txid)]; !ok || val.ID == nil {
			return ErrIncorrectTransaction
		}
	}
	txCopy := tx.TrimmedCopy()
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		if txCopy.ID, err = txCopy.Hash(); err != nil {
			return err
		}
		txCopy.Vin[inID].PubKey = nil
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			return err
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}
	return nil
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

func (tx *Transaction) Hash() ([]byte, error) {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	serialized, err := txCopy.Serialize()
	if err != nil {
		return nil, err
	}
	hash = sha256.Sum256(serialized)
	return hash[:], nil
}

func (tx *Transaction) Serialize() ([]byte, error) {
	var encoded bytes.Buffer
	if err := gob.NewEncoder(&encoded).Encode(tx); err != nil {
		return nil, err
	}
	return encoded.Bytes(), nil
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) (bool, error) {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	var err error

	for inID, vin := range tx.Vin {
		prevTx, ok := prevTXs[hex.EncodeToString(vin.Txid)]
		if !ok {
			return false, nil
		}
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		if txCopy.ID, err = txCopy.Hash(); err != nil {
			return false, err
		}
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false, nil
		}
	}
	return true, nil
}
