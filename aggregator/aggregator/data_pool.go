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

package aggregator

import evmCommon "github.com/ethereum/go-ethereum/common"

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
