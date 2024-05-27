package aggregator

import (
	"testing"

	evmCommon "github.com/ethereum/go-ethereum/common"
)

func TestDataSet(t *testing.T) {
	ds := NewDataSet()

	ds.Add(evmCommon.BytesToHash([]byte("0_1")), "0_1", 0)
	ds.Add(evmCommon.BytesToHash([]byte("1_1")), "1_1", 1)
	ds.Add(evmCommon.BytesToHash([]byte("1_3")), "1_3", 1)
	ds.Add(evmCommon.BytesToHash([]byte("2_1")), "2_1", 2)
	list := ds.Get([]evmCommon.Hash{
		evmCommon.BytesToHash([]byte("1_1")),
		evmCommon.BytesToHash([]byte("1_2")),
	}, 1)
	if list != nil {
		t.Fail()
	}

	list = ds.Get([]evmCommon.Hash{
		evmCommon.BytesToHash([]byte("1_2")),
		evmCommon.BytesToHash([]byte("1_3")),
	}, 1)
	if list != nil {
		t.Fail()
	}

	lists := ds.Add(evmCommon.BytesToHash([]byte("1_2")), "1_2", 1)
	t.Log(lists)

	ds.Add(evmCommon.BytesToHash([]byte("2_2")), "2_2", 2)
	list = ds.Get([]evmCommon.Hash{
		evmCommon.BytesToHash([]byte("2_1")),
		evmCommon.BytesToHash([]byte("2_2")),
	}, 2)
	t.Log(list)

	ds.Clear(1)
	t.Log(ds)
}
