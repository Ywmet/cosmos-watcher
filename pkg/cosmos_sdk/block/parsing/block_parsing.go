package cosmos

import (
	"bytes"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/x/auth"
	types "github.com/mapofzones/cosmos-watcher/pkg/cosmos_sdk/block/types"
	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
	"github.com/tendermint/go-amino"
)

func txToMessage(tx auth.StdTx, hash string, errCode uint32) (watcher.Message, error) {
	Tx := watcher.Transaction{
		Hash:     hash,
		Accepted: errCode == 0,
	}
	for _, msg := range tx.Msgs {
		msgs, err := parseMsg(msg)
		if err != nil {
			return Tx, err
		}
		Tx.Messages = append(Tx.Messages, msgs...)
	}
	return Tx, nil
}

func txErrCode(b types.Block, hash []byte) uint32 {
	for _, res := range b.Results {
		if bytes.Equal(res.Hash, hash) {
			return res.ResultCode
		}
	}

	panic("could not find tx status for given tx hash")
}

func DecodeBlock(cdc *amino.Codec, b types.Block) (types.ProcessedBlock, error) {
	block := types.ProcessedBlock{
		Height_:          b.Height,
		ChainID_:         b.ChainID,
		BeginBlockEvents: nil,
		EndBlockEvents:   nil,
		T:                b.T,
	}

	block.Txs = make([]watcher.Message, 0, len(b.Txs))
	for _, tx := range b.Txs {
		decoded, err := decodeTx(cdc, tx)
		if err != nil {
			return block, err
		}

		txMessage, err := txToMessage(decoded, hex.EncodeToString(tx.Hash()), txErrCode(b, tx.Hash()))
		if err != nil {
			return block, err
		}
		block.Txs = append(block.Txs, txMessage)
	}

	return block, nil
}
