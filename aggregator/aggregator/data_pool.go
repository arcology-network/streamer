package aggregator

import evmCommon "github.com/arcology-network/evm/common"

type DataPool struct {
	data map[evmCommon.Hash]interface{}
}

// NewDataPool returns a new DataPool structure.
func NewDataPool() *DataPool {
	return &DataPool{
		data: map[evmCommon.Hash]interface{}{},
	}
}

// object and raw enter pool
func (d *DataPool) add(h evmCommon.Hash, data interface{}) {
	d.data[h] = data
}

// get data and raw
func (d *DataPool) get(h evmCommon.Hash) interface{} {
	return d.data[h]
}

// remove data and raw from pool
func (d *DataPool) remove(h evmCommon.Hash) {
	delete(d.data, h)
}

// remove data and raw from pool
func (d *DataPool) count() int {
	return len(d.data)
}

// range data_pool
func (d *DataPool) Range(f func(hash evmCommon.Hash, val interface{}) bool) {
	for k, v := range d.data {
		if !f(k, v) {
			break
		}
	}
}
