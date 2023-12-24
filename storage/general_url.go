package storage

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/transactional"

	cmntyp "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	intf "github.com/arcology-network/component-lib/interface"
	"github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
	ccdb "github.com/arcology-network/concurrenturl/storage"
)

type GeneralUrl struct {
	BasicDBOperation

	generateApcHandle bool
	generateUrlUpdate bool
	inited            bool
	cached            bool
	apcHandleName     string
}

type UrlUpdate struct {
	Keys          []string
	EncodedValues [][]byte
}

func NewGeneralUrl(apcHandleName string) *GeneralUrl {
	return &GeneralUrl{
		apcHandleName: apcHandleName,
	}
}

func (url *GeneralUrl) PreCommit(euResults []*cmntyp.EuResult, height uint64) {
	url.BasicDBOperation.PreCommit(euResults, height)
	if url.generateUrlUpdate {
		keys, values := url.URL.KVs()
		keys = codec.Strings(keys).Clone()
		encodedValues := make([][]byte, len(values))
		metaKeys := make([]string, len(keys))
		encodedMetas := make([][]byte, len(keys))
		worker := func(start, end, index int, args ...interface{}) {
			for i := start; i < end; i++ {
				if values[i] != nil {
					univalue := values[i].(interfaces.Univalue)
					if univalue.Value() != nil && univalue.Preexist() && univalue.Value().(interfaces.Type).TypeID() == commutative.PATH { // Skip meta data
						metaKeys[i] = keys[i]
						encodedMetas[i] = ccdb.Codec{}.Encode("", univalue.Value()) //urltyp.ToBytes(univalue.Value())

						keys[i] = ""
						continue
					}
					encodedValues[i] = ccdb.Codec{}.Encode("", univalue.Value()) //urltyp.ToBytes(univalue.Value())
				} else {
					encodedValues[i] = nil
				}
			}
		}
		common.ParallelWorker(len(keys), 4, worker)

		// common.RemoveEmptyStrings(&keys)
		// common.RemoveNilBytes(&encodedValues)
		// common.RemoveEmptyStrings(&metaKeys)
		// common.RemoveNilBytes(&encodedMetas)

		filter := func(v []byte) bool { return v == nil }
		common.Remove(&keys, "")
		common.RemoveIf(&encodedValues, filter)
		common.Remove(&metaKeys, "")
		common.RemoveIf(&encodedMetas, filter)

		var na int
		if len(keys) > 0 {
			intf.Router.Call("transactionalstore", "AddData", &transactional.AddDataRequest{
				Data: &UrlUpdate{
					Keys:          keys,
					EncodedValues: encodedValues,
				},
				RecoverFunc: "urlupdate",
			}, &na)
		}
		if len(metaKeys) > 0 {
			intf.Router.Call("transactionalstore", "AddData", &transactional.AddDataRequest{
				Data: &UrlUpdate{
					Keys:          metaKeys,
					EncodedValues: encodedMetas,
				},
				RecoverFunc: "urlupdate",
			}, &na)
		}

		url.MsgBroker.Send(actor.MsgUrlUpdate, &UrlUpdate{
			Keys:          keys,
			EncodedValues: encodedValues,
		})
	}
}

func (url *GeneralUrl) Finalize() {
	if !url.inited {
		url.inited = true
	} else {
		url.BasicDBOperation.Finalize()
	}

	if url.generateApcHandle {
		url.MsgBroker.Send(url.apcHandleName, &url.DB)
	}

	if url.cached {
		url.MsgBroker.Send(actor.MsgCached, "")
	}
}

func (url *GeneralUrl) Outputs() map[string]int {
	outputs := make(map[string]int)
	if url.generateApcHandle {
		outputs[url.apcHandleName] = 1
	}
	if url.generateUrlUpdate {
		outputs[actor.MsgUrlUpdate] = 1
	}
	if url.cached {
		outputs[actor.MsgCached] = 1
	}
	return outputs
}

func (url *GeneralUrl) Config(params map[string]interface{}) {
	if v, ok := params["generate_apc_handle"]; !ok {
		panic("parameter not found: generate_apc_handle")
	} else {
		url.generateApcHandle = v.(bool)
	}

	if v, ok := params["generate_url_update"]; !ok {
		panic("parameter not found: generate_url_update")
	} else {
		url.generateUrlUpdate = v.(bool)
	}
	if v, ok := params["cached"]; !ok {
		panic("parameter not found: cached")
	} else {
		url.cached = v.(bool)
	}
}
