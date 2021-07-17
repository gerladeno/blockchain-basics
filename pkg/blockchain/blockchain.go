package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"github.com/boltdb/bolt"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type BCIterator struct {
	currentHash []byte
	db          *bolt.DB
}

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func GetBlockchain() (*Blockchain, error) {
	if !dbExists() {
		return nil, errors.New("db doesn't exist, create it first")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		return nil, err
	}
	bc := Blockchain{tip: tip, db: db}
	return &bc, nil
}

func CreateBlockchain(address string) (*Blockchain, error) {
	var tip, serialized []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer func(db *bolt.DB) {
		if err = db.Close(); err != nil {
			panic(err)
		}
	}(db)
	var cbtx *Transaction
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			cbtx, err = CreateCoinbaseTX(address, genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)
			b, err = tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}
			if serialized, err = genesis.Serialize(); err != nil {
				return err
			}
			if err = b.Put(genesis.Hash, serialized); err != nil {
				return err
			}
			if err = b.Put([]byte("l"), genesis.Hash); err != nil {
				return err
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	bc := Blockchain{tip, db}
	return &bc, err
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) error {
	var lastHash, serialized []byte
	var ok bool
	var err error
	for _, tx := range transactions {
		if ok, err = bc.VerifyTransaction(tx); err != nil {
			return err
		}
		if !ok {
			return ErrIncorrectTransaction
		}
	}

	err = bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return err
	}

	newBlock := NewBlock(transactions, lastHash)
	if serialized, err = newBlock.Serialize(); err != nil {
		return err
	}

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if err = b.Put(newBlock.Hash, serialized); err != nil {
			return err
		}
		if err = b.Put([]byte("l"), newBlock.Hash); err != nil {
			return err
		}
		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) Iterator() *BCIterator {
	return &BCIterator{currentHash: bc.tip, db: bc.db}
}

func (bci *BCIterator) Next() (*Block, error) {
	var block *Block
	var err error

	err = bci.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(bci.currentHash)
		block, err = Deserialize(encodedBlock)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	bci.currentHash = block.PrevBlockHash

	return block, nil
}

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) ([]Transaction, error) {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	var usesKey bool
	for {
		block, err := bci.Next()
		if err != nil {
			return nil, err
		}
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIDx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIDx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					usesKey, err = in.UsesKey(pubKeyHash)
					if err != nil {
						return nil, err
					}
					if usesKey {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs, nil
}

func (bc *Blockchain) FindUTXO(pubKeyHash []byte) ([]TXOutput, error) {
	var UTXOs []TXOutput
	unspentTransactions, err := bc.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return nil, err
	}
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs, nil
}

func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOutputs := make(map[string][]int)
	unspentTXs, err := bc.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return 0, nil, err
	}
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs, nil
}

func (bc *Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			return nil, err
		}
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil, ErrTransactionNotFound
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return err
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}
	if err := tx.Sing(privKey, prevTXs); err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) (bool, error) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return false, err
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}
	return tx.Verify(prevTXs)
}
