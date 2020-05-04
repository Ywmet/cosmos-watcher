package block

import (
	"github.com/tendermint/tendermint/types"
)

// Block is like tendermint block but without unnecessary mutex and unused data
type Block struct {
	types.Header
	types.Txs
}

// BlockFromTmResultBlock parses tm websocket response to fit our block structure
func BlockFromTmResultBlock(b types.EventDataNewBlock) Block {
	return Block{
		Header: b.Block.Header,
		Txs:    b.Block.Txs,
	}
}

// TxStatus only stores transaction result code and it's hash
type TxStatus struct {
	ResultCode uint32
	Hash       []byte
	Height     int64
}

// TxStatusFromTmResultTx parses ResultTx to give us tx hash, height and error code
// actual tx body is stored in the block
func TxStatusFromTmResultTx(t types.EventDataTx) TxStatus {
	return TxStatus{
		ResultCode: t.Result.Code,
		Hash:       t.Tx.Hash(),
		Height:     t.Height,
	}
}
