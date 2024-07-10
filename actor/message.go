/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package actor

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/big"
	"sync"

	"github.com/arcology-network/common-lib/common"
	evmCommon "github.com/ethereum/go-ethereum/common"
)

const (
	MsgGeneralDB                  = "generalDB"
	MsgNonceDB                    = "nonceDB"
	MsgGeneralCompleted           = "generalCompleted"
	MsgNonceCompleted             = "nonceCompleted"
	MsgGeneralPrecommit           = "generalPrecommit"
	MsgNoncePrecommit             = "noncePrecommit"
	MsgGeneralCommit              = "generalCommit"
	MsgNonceCommit                = "nonceCommit"
	MsgInitDBGeneral              = "initdbGeneral"
	MsgInitDBNonce                = "initdbNonce"
	MsgInclusivePre               = "inclusivePre"
	MsgInclusive                  = "inclusive"
	MsgGenerationReapingListPre   = "generationReapingListPre"
	MsgGenerationReapingList      = "generationReapingList"
	MsgGenerationReapingCompleted = "generationReapingCompleted"
	MsgConflictInclusive          = "conflictInclusive"
	MsgBlockCompleted             = "blockCompleted"
	MsgTxHash                     = "txhash"
	MsgRcptHash                   = "rcpthash"
	MsgAcctHash                   = "accthash"
	MsgGasUsed                    = "gasused"
	MsgBloom                      = "bloom"
	MsgTpsGasBurned               = "tpsGasBurned"
	MsgParentInfo                 = "parentinfo"
	MsgLocalParentInfo            = "localparentinfo"
	MsgCheckedTxs                 = "checkedTxs" // common-lib/types.*IncomingTxs
	MsgTxBlocks                   = "txBlocks"   // common-lib/types.*IncomingTxs
	MsgTxLocals                   = "txLocals"
	MsgTxLocalsUnChecked          = "txLocalsunchecked" // main/modules/gateway/types.*TxsPack
	MsgMetaBlock                  = "metablock"
	MsgSelectedTx                 = "selectedtx"
	MsgPendingBlock               = "pendingblock"
	MsgExecTime                   = "execTime"
	MsgReceiptHashList            = "receiptHashList"
	MsgEuResults                  = "euResults"
	MsgNonceEuResults             = "nonceEuResults"
	MsgTxAccessRecords            = "txAccessRecords"
	MsgExecutingLogs              = "executingLogs"
	MsgPreProcessedEuResults      = "preProcessedEuResults"
	MsgApcHandleInit              = "apchandleInit"
	MsgApcHandle                  = "apchandle"
	MsgNonceReady                 = "nonceready"
	MsgCached                     = "cached"
	MsgObjectCached               = "objectCached"
	MsgChainConfig                = "chainConfig"
	MsgUrlUpdate                  = "urlupdate"
	MsgTransactionalAddCompleted  = "transactionalAddCompleted"
	MsgSchdState                  = "schdstate"
	MsgReapinglist                = "reapinglist"
	MsgArbitrateReapinglist       = "arbitratereapinglist"
	MsgExecuted                   = "executed"
	MsgCommitNonceUrl             = "commitNonceUrl"
	MsgSelectedExecuted           = "selectedexecuted"
	MsgListFulfilled              = "listfulfilled"
	MsgApcBlock                   = "apcBlock"
	MsgSelectedReceipts           = "selectedReceipts"
	MsgSelectedReceiptsHash       = "selectedReceiptsHash"
	MsgReceipts                   = "receipts"
	MsgCheckingTxs                = "checkingtxs" // main/modules/tpp/types.*CheckingTxsPack
	MsgMessager                   = "messager"    // common-lib/types.*IncomingMsgs
	MsgBlockCompleted_Success     = "success"
	MsgBlockCompleted_Failed      = "failed"
	MsgMessagersReaped            = "messagersReaped"
	MsgArbitrateList              = "arbitrateList"
	MsgTxsToExecute               = "txsToExecute"      // common-lib/types.*ExecutorRequest
	MsgTxsExecuteResults          = "txsExecuteResults" // main/modules/exec/v2.[]*ExecutorResponse
	MsgEuResultSelected           = "euResultSelected"
	MsgTxs                        = "txs"
	MsgPrecedingList              = "precedingList"      // 3rd-party/eth/common.*[]*Hash
	MsgPrecedingsEuresult         = "precedingsEuresult" // []interface{} (common-lib/types.[]*EuResult)
	MsgReapCommand                = "reapCommand"
	MsgOpCommand                  = "opCommand"
	MsgBlockParams                = "blockParams"
	MsgWithDrawHash               = "withDrawHash"
	MsgSignerType                 = "signerType"
	MsgAppHash                    = "appHash"
	MsgSpawnedRelations           = "spawnedRelations"
	MsgGc                         = "gc"

	MsgInitDB = "initdb"

	MsgBlockStart             = "blockstart"
	MsgBlockEnd               = "blockend"
	MsgStorageUp              = "storage.up"
	MsgCoinbase               = "coinbase"
	MsgFastSyncDone           = "storage.fastsyncdone"
	MsgConsensusMaxPeerHeight = "consensus.maxpeerheight"
	MsgConsensusUp            = "consensus.up"

	MsgExtBlockStart     = "external.blockstart"
	MsgExtBlockEnd       = "external.blockend"
	MsgExtReapCommand    = "external.reapcommand"
	MsgExtTxBlocks       = "external.txblocks" // common-lib/types.*IncomingTxs
	MsgExtBlockCompleted = "external.blockcompleted"
	MsgExtReapingList    = "external.reapinglist"
	MsgExtAppHash        = "external.apphash"

	MsgStateSyncStart = "statesync.start"
	MsgStateSyncDone  = "statesync.done"

	MsgP2pRequest         = "p2p.request"
	MsgP2pResponse        = "p2p.response"
	MsgSyncStatusRequest  = "statesync.status.request"
	MsgSyncStatusResponse = "statesync.status.response"
	MsgSyncPointRequest   = "statesync.sp.request"
	MsgSyncPointResponse  = "statesync.sp.request"
	MsgSyncDataRequest    = "statesync.data.request"
	MsgSyncDataResponse   = "statesync.data.response"
	MsgSyncTxRequest      = "txsync.request"
	MsgSyncTxResponse     = "txsync.response"

	MsgP2pReceived = "p2p.received"
	MsgP2pSent     = "p2p.sent"
)

// BlockStart used with MsgBlockStart.
type BlockStart struct {
	Timestamp *big.Int
	Coinbase  evmCommon.Address
	Height    uint64
	Extra     []byte
}

type Comparable interface {
	Equals(rhs Comparable) bool
}

func init() {
	gob.Register(&Message{})
}

type Message struct {
	From        string      `json:"from"`
	Msgid       uint64      `json:"msgid"`
	Name        string      `json:"name"`
	Height      uint64      `json:"height"`
	Round       uint64      `json:"round"`
	Data        interface{} `json:"data"`
	encodedSize uint32

	lock       sync.Mutex
	isReadOnly bool
	owner      string
	readers    []string
}

func NewMessage() *Message {
	return &Message{
		Msgid:  0,
		Name:   "",
		Height: 0,
		Round:  0,
		Data:   nil,
	}
}

func (m *Message) Read(reader string) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.isReadOnly || (!m.isReadOnly && len(m.owner) == 0) {
		m.readers = append(m.readers, reader)
	} else {
		panic(fmt.Sprintf("%s tries to read message[%s], but the message is already owned by %s", reader, m.Name, m.owner))
	}
	return m.Data
}

func (m *Message) ReadCopy(reader string) interface{} {
	data := m.Read(reader)
	if copyable, ok := data.(Copyable); ok {
		return copyable.Clone()
	}
	panic(fmt.Sprintf("%s tries to read copy of message[%s], but the message data is not copyable", reader, m.Name))
}

func (m *Message) TakeOver(owner string) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.isReadOnly {
		panic(fmt.Sprintf("%s tries to take over message[%s], but the message was declared readonly by %s", owner, m.Name, m.owner))
	} else if len(m.owner) != 0 {
		panic(fmt.Sprintf("%s tries to take over message[%s], but the message is already owned by %s", owner, m.Name, m.owner))
	} else if len(m.readers) != 0 {
		panic(fmt.Sprintf("%s tries to take over message[%s], but the message is already read by %v", owner, m.Name, m.readers))
	}
	m.owner = owner
	return m.Data
}

func (m *Message) Equals(rhs Comparable) bool {
	other := rhs.(*Message)
	return m.Name == other.Name && m.Height == other.Height && m.Round == other.Round && m.Data == other.Data
}

func (m *Message) CopyHeader() *Message {
	return &Message{
		From:        m.From,
		Msgid:       m.Msgid,
		Name:        m.Name,
		Height:      m.Height,
		Round:       m.Round,
		encodedSize: m.encodedSize,
	}
}

func (msg *Message) Encode() ([]byte, error) {
	data, err := common.GobEncode(&msg)
	msg.encodedSize = uint32(len(data))
	return data, err
}

func (msg *Message) Decode(data []byte) error {
	msg.encodedSize = uint32(len(data))
	return common.GobDecode(data, &msg)
}

func (msg *Message) Size() uint32 {
	return msg.encodedSize
}

func (msg *Message) GetHeader() (uint64, uint64, uint64) {
	return msg.Height, msg.Round, msg.Msgid
}

func (msg *Message) Hash() evmCommon.Hash {
	hash := evmCommon.Hash{}
	binary.LittleEndian.PutUint64(hash[:], uint64(msg.Msgid))
	return hash
}

type MessageCache struct {
	height      uint64
	msgs        map[uint64][]*Message
	executeFunc ExecuteMsg
}

func NewMessageCache(fun ExecuteMsg) *MessageCache {
	return &MessageCache{
		msgs:        map[uint64][]*Message{},
		executeFunc: fun,
		height:      0,
	}
}

func (cache *MessageCache) TryExecute(msg *Message) bool {
	if msg.Height == cache.height {
		cache.executeFunc(msg)
		return true
	}
	ms := cache.msgs[msg.Height]
	ms = append(ms, msg)
	cache.msgs[msg.Height] = ms
	return false
}

func (cache *MessageCache) ChangeHeight(height uint64) {
	cache.height = height
	for _, msg := range cache.msgs[height] {
		cache.executeFunc(msg)
	}
	delete(cache.msgs, height)
}

type ExecuteMsg func(msg *Message) error
