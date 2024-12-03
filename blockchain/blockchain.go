package blockchain

import (
	"encoding/hex"
	"fmt"
	"iter"
	"os"
	"runtime"

	"github.com/dgraph-io/badger/v4"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash    []byte
	CurrentHash []byte
	Database    *badger.DB
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func ContinueBlockChain(address string) *BlockChain {
	if !DBExists() {
		fmt.Println("No existing blockchain found. Create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.ValueDir = dbPath
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		return err
	})
	Handle(err)

	chain := BlockChain{LastHash: lastHash, CurrentHash: lastHash, Database: db}
	return &chain
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists() {
		fmt.Println("Blockchain already exists.")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	opts.ValueDir = dbPath
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			cbtx := CoinbaseTx(address, genesisData)
			genesis := Genesis(cbtx)
			fmt.Println("Genesis created")
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

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lashHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error {
			lashHash = val
			return nil
		})
		return err
	})
	Handle(err)

	newBlock := CreateBlock(transactions, lashHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
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
		currentHash := chain.LastHash

		for {
			var block *Block
			// ブロックの取得
			err := chain.Database.View(func(txn *badger.Txn) error {
				item, err := txn.Get(currentHash)
				if err != nil {
					return err
				}

				return item.Value(func(val []byte) error {
					block = Deserialize(val)
					return nil
				})
			})

			if err != nil {
				break
			}

			// イテレータを継続するかチェック
			if !yield(block) {
				break
			}

			// 次のブロックへ
			currentHash = block.PrevHash

			// ジェネシスブロックに到達したら終了
			if len(currentHash) == 0 {
				break
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

func (chain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	iter(func(block *Block) bool {
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		return len(block.PrevHash) != 0
	})

	return unspentTxs
}

func (chain *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// 指定されたアドレスから指定金額分の使用可能なUTXOを探す
//
// params:
//   - address: UTXOを探すアドレス
//   - amount: 必要な合計金額
//
// returns:
//   - int: 見つかった未使用出力の合計金額
//   - map[string][]int: トランザクションIDをキーとし、使用可能な出力インデックスの配列を値とするマップ
func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
