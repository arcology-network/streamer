package storage

import (
	"crypto/sha256"
	"sync"
	"time"

	"github.com/arcology-network/component-lib/actor"
	kafkalib "github.com/arcology-network/component-lib/kafka/lib"
	"github.com/arcology-network/component-lib/log"
	"go.uber.org/zap"
)

type CheckedListItem struct {
	createTime time.Time
	tx         []byte
	from       byte
}

type CheckedList struct {
	all             map[[32]byte]CheckedListItem
	lock            sync.RWMutex
	totals          uint64
	hits            uint64
	timeoutForClear time.Duration
}

// NewList returns a new CheckedList structure.
func NewCheckList(waits int64) *CheckedList {
	cl := CheckedList{
		all:             make(map[[32]byte]CheckedListItem),
		totals:          0,
		hits:            0,
		timeoutForClear: time.Duration(waits) * time.Second,
	}
	tim := kafkalib.SyncTimer{}
	tim.StartTimer(cl.timeoutForClear, cl.timerClear)
	return &cl
}

func (t *CheckedList) timerClear() {
	t.lock.Lock()
	defer t.lock.Unlock()

	clearCounter := 0
	for k, v := range t.all {
		if v.createTime.Add(t.timeoutForClear).Before(time.Now()) {
			delete(t.all, k)
			clearCounter++
		}
	}
	log.Logger.AddLog(log.Logger.GetLogId(), log.LogLevel_Debug, "checklist", "unsigner", "checklist hit rate", log.LogType_Inlog, 0, 0, 0, 0, zap.Uint64("checked", t.totals), zap.Uint64("hit", t.hits))
}

func (t *CheckedList) ExistTx(tx []byte, from byte, inlog *actor.WorkerThreadLogger) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.totals = t.totals + 1
	hash := sha256.Sum256(tx)
	_, ok := t.all[hash]
	if !ok {
		t.all[hash] = CheckedListItem{
			createTime: time.Now(),
			tx:         tx,
			from:       from,
		}
		return false
	}
	t.hits = t.hits + 1
	//inlog.Log(log.LogLevel_Debug, "checkingTxs repeated", zap.Uint64("hit", t.hits), zap.Time("firstTime", first.createTime), zap.Int8("firstFrom", int8(first.from)), zap.Int8("txFrom", int8(from)), zap.String("tx", fmt.Sprintf("%x", tx)))
	return true
}
