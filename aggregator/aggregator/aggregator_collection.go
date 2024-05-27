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

type AggregateCollection struct {
	accumulated []interface{}
}

func NewAggregateCollection() *AggregateCollection {
	return &AggregateCollection{
		accumulated: []interface{}{},
	}
}

// action when clear received
func (ac *AggregateCollection) OnReset() {
	ac.accumulated = []interface{}{}
}

// action when new batch received
func (ac *AggregateCollection) OnNewBatchReceived(addBatch *[]interface{}) *[]interface{} {
	for _, v := range *addBatch {
		ac.accumulated = append(ac.accumulated, v)
	}
	return &ac.accumulated
}

// get all items
func (ac *AggregateCollection) GetAll() *[]interface{} {
	return &ac.accumulated
}

func (ac *AggregateCollection) ConvertToHashList(raws *[]interface{}) *[]*evmCommon.Hash {
	if raws == nil || len(*raws) == 0 {
		return &[]*evmCommon.Hash{}
	}
	rets := make([]*evmCommon.Hash, len(*raws))
	for i, v := range *raws {
		rets[i] = v.(*evmCommon.Hash)
	}
	return &rets
}
