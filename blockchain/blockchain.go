package blockchain

import (
	"fmt"
	"iter"

	"github.com/dgraph-io/badger/v4"
)

const (
	dbPath = "./tmp/blocks"
)

type BlockChain struct {
	LastHash    []byte
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.ValueDir = dbPath
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			err = item.Value(func(val []byte) error {
				lastHash = val
				return nil
			})
			return err
		}
	})

	Handle(err)

	blockchain := BlockChain{LastHash: lastHash, CurrentHash: lastHash, Database: db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	newBlock := CreateBlock(data, chain.LastHash)

	err := chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}

func (chain *BlockChain) Iterator() iter.Seq[*Block] {
	return func(yield func(*Block) bool) {
		chain.CurrentHash = chain.LastHash
		for {
			block := chain.Next()
			if block == nil {
				return
			}
			if !yield(block) {
				return
			}
		}
	}
}

func (chain *BlockChain) Next() *Block {
	var block *Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(chain.CurrentHash)
		Handle(err)
		err = item.Value(func(val []byte) error {
			encodedBlock := val
			block = Deserialize(encodedBlock)
			return nil
		})

		return err
	})
	Handle(err)

	chain.CurrentHash = block.PrevHash

	return block
}
