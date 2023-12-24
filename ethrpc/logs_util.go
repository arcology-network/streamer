package ethrpc

import (
	eth "github.com/arcology-network/evm"
	ethcmn "github.com/arcology-network/evm/common"
	ethtyp "github.com/arcology-network/evm/core/types"
)

type LogCache struct {
	Logs      []*ethtyp.Log
	Height    uint64
	BlockHash ethcmn.Hash
}

func containAddress(addrs []ethcmn.Address, addr ethcmn.Address) bool {
	for _, ad := range addrs {
		if ad == addr {
			return true
		}
	}
	return false
}
func FiltereTopic(tpoicsFilter [][]ethcmn.Hash, tpoics []ethcmn.Hash) bool {
	if len(tpoicsFilter) > len(tpoics) {
		return false
	}
	for i := range tpoicsFilter {
		if len(tpoicsFilter[i]) == 0 {
			continue
		}
		found := false
		for j := range tpoicsFilter[i] {
			if tpoicsFilter[i][j] == tpoics[i] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func FilteLogs(logs []*ethtyp.Log, filter eth.FilterQuery) []*ethtyp.Log {
	fielteredLogs := make([]*ethtyp.Log, 0, len(logs))
	if filter.Addresses == nil || len(filter.Addresses) == 0 {
		fielteredLogs = logs
	} else {
		for _, log := range logs {
			if containAddress(filter.Addresses, log.Address) {
				fielteredLogs = append(fielteredLogs, log)
			}
		}
	}
	topicFielteredLogs := make([]*ethtyp.Log, 0, len(fielteredLogs))
	if filter.Topics == nil || len(filter.Topics) == 0 {
		topicFielteredLogs = fielteredLogs
	} else {
		for _, log := range fielteredLogs {
			if FiltereTopic(filter.Topics, log.Topics) {
				topicFielteredLogs = append(topicFielteredLogs, log)
			}
		}
	}
	return topicFielteredLogs
}
func ToLogs(receipts []*ethtyp.Receipt) []*ethtyp.Log {
	logsSize := 0
	for i := range receipts {
		logsSize += len(receipts[i].Logs)
	}
	logs := make([]*ethtyp.Log, 0, logsSize)

	for _, receipt := range receipts {
		logs = append(logs, receipt.Logs...)
		// for _, log := range receipt.Logs {
		// 	topics := make([]ethcmn.Hash, len(log.Topics))
		// 	for j, topic := range log.Topics {
		// 		topics[j] = ethcmn.BytesToHash(topic.Bytes())
		// 	}
		// 	evmlog := &ethtyp.Log{
		// 		Address:     ethcmn.BytesToAddress(log.Address.Bytes()),
		// 		Topics:      topics,
		// 		Data:        log.Data,
		// 		BlockNumber: log.BlockNumber,
		// 		TxHash:      ethcmn.BytesToHash(log.TxHash.Bytes()),
		// 		TxIndex:     log.TxIndex,
		// 		BlockHash:   ethcmn.BytesToHash(log.BlockHash.Bytes()),
		// 		Index:       log.Index,
		// 		Removed:     log.Removed,
		// 	}
		// 	logs = append(logs, evmlog)
		// }
	}
	return logs
}
