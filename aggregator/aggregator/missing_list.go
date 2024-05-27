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

import (
	"sync"

	evmCommon "github.com/ethereum/go-ethereum/common"
)

type MissingList struct {
	missingList map[evmCommon.Hash]int
	isGenerated bool
	mtx         sync.Mutex
}

// NewList returns a new MissingList structure.
func NewMissingList() *MissingList {
	return &MissingList{
		missingList: map[evmCommon.Hash]int{},
		isGenerated: false,
	}
}

// received is completed or not
func (m *MissingList) IsReceiveCompleted() bool {
	if m.isGenerated && len(m.missingList) == 0 {
		return true
	} else {
		return false
	}
}

// clear missinglist
func (m *MissingList) ClearList() {
	m.missingList = map[evmCommon.Hash]int{}
	m.isGenerated = false
}

// set missinglist init terminated
func (m *MissingList) CompleteGnereation() {
	m.isGenerated = true
}

// init terminated or not
func (m *MissingList) IsGenerated() bool {
	return m.isGenerated
}

// put a element into missinglist
func (m *MissingList) Put(hash evmCommon.Hash, idx int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.missingList[hash] = idx

}

// remove a element from missinglist
func (m *MissingList) RemoveFromMissing(hash evmCommon.Hash) (bool, int) {
	if idx, ok := m.missingList[hash]; ok {
		delete(m.missingList, hash)
		return true, idx
	}
	return false, -1
}

// remove a element from missinglist
func (m *MissingList) Remove(hash evmCommon.Hash) {
	delete(m.missingList, hash)
}

// the hash in missings or not
func (m *MissingList) IsExist(hash evmCommon.Hash) (bool, int) {
	if idx, ok := m.missingList[hash]; ok {
		return true, idx
	}
	return false, -1
}

// missings size
func (m *MissingList) Size() int {
	return len(m.missingList)
}

// range missinglist
func (m *MissingList) Range(f func(hash evmCommon.Hash, idx int) bool) {
	for k, v := range m.missingList {
		if f(k, v) {
			delete(m.missingList, k)
		}
	}
}
