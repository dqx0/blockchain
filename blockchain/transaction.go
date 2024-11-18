package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte     // トランザクションのID
	Inputs  []TxInput  // 入力トランザクションのリスト
	Outputs []TxOutput // 出力トランザクションのリスト
}

type TxOutput struct {
	Value  int    // 送金額
	PubKey string // 受取人の公開鍵
}

type TxInput struct {
	ID  []byte // 参照する過去のトランザクションID
	Out int    // 参照する出力インデックス
	Sig string // 送信者の署名
}

// トランザクションのハッシュIDを生成
// トランザクションの内容をgobでエンコードし、SHA-256ハッシュを計算
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// マイニング報酬用の新規コインを生成するトランザクションを作成
// params:
//
//	to: 報酬を受け取るアドレス
//	data: カスタムデータ
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	// コインベーストランザクションは過去の参照を持たない特殊なインプット
	txin := TxInput{[]byte{}, -1, data}
	// 報酬として100コインを設定
	txout := TxOutput{100, to}
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetID()

	return &tx
}

// 通常の送金トランザクションを作成
// params:
//
//	from: 送金元アドレス
//	to: 送金先アドレス
//	amount: 送金額
//	bc: ブロックチェーンインスタンス
func NewTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	// 利用可能なUTXOを検索
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: Not enough funds")
	}

	// 入力トランザクションの作成
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// 出力トランザクションの作成
	outputs = append(outputs, TxOutput{amount, to})
	// おつりがある場合は送金元に返す
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

// トランザクションがコインベース（マイニング報酬）かどうかを判定
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// 指定されたデータで入力トランザクションがアンロック可能か検証
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// 指定されたデータで出力トランザクションが使用可能か検証
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
