package blockchain

type TxOutput struct {
	Value  int    // 送金額
	PubKey string // 受取人の公開鍵
}

type TxInput struct {
	ID  []byte // 参照する過去のトランザクションID
	Out int    // 参照する出力インデックス
	Sig string // 送信者の署名
}

// 指定されたデータで入力トランザクションがアンロック可能か検証
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// 指定されたデータで出力トランザクションが使用可能か検証
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
