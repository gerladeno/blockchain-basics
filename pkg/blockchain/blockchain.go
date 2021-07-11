package blockchain

import "github.com/boltdb/bolt"

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
)

type BCIterator struct {
	currentHash []byte
	db          *bolt.DB
}

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func NewBlockchain() (*Blockchain, error) {
	var tip, serialized []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock()
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

func (bc *Blockchain) AddBlock(data string) error {
	var lastHash, serialized []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return err
	}

	newBlock := NewBlock(data, lastHash)
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
