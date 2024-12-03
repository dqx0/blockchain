package blockchain

import (
	"bytes"

	"github.com/dqx0/blockchain/wallet"
)

type TxOutput struct {
	Value      int    // 送金額
	PubKeyHash []byte // 受取人の公開鍵
}

type TxInput struct {
	ID     []byte // 参照する過去のトランザクションID
	Out    int    // 参照する出力インデックス
	Sig    []byte // 送信者の署名
	PubKey []byte // 送信者の公開鍵
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
