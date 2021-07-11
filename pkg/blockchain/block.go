package blockchain

import (
	"bytes"
	"encoding/gob"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
	pow := NewProofOfWork(&block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return &block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

func (block *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer
	if err := gob.NewEncoder(&result).Encode(block); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func Deserialize(data []byte) (*Block, error) {
	var block Block
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&block); err != nil {
		return nil, err
	}
	return &block, nil
}
