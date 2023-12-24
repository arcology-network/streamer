package ethrpc

import (
	ethcmn "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/common/hexutil"
	ethtyp "github.com/arcology-network/evm/core/types"
)

type RPCTransaction struct {
	BlockHash        *ethcmn.Hash    `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             ethcmn.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             ethcmn.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *ethcmn.Address `json:"to"`
	TransactionIndex *hexutil.Uint64 `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	Type             hexutil.Uint64  `json:"type"`
	ChainID          *hexutil.Big    `json:"chainId,omitempty"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

type RPCBlock struct {
	Header       *ethtyp.Header
	Transactions []interface{}
}

type FeeHistoryResult struct {
	OldestBlock  *hexutil.Big     `json:"oldestBlock"`
	Reward       [][]*hexutil.Big `json:"reward,omitempty"`
	BaseFee      []*hexutil.Big   `json:"baseFeePerGas,omitempty"`
	GasUsedRatio []float64        `json:"gasUsedRatio"`
}

const (
	BlockNumberLatest   = -1
	BlockNumberEarliest = -2
	BlockNumberPending  = -3
)
