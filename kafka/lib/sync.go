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

package lib

import (
	"sync"
	"time"
)

// type Waitobjs map[int64]WaitObject
type Waitobjs struct {
	Wbs map[int64]*WaitObject
	mtx sync.Mutex
}

// start a wait objects
func StartWaitObjects() *Waitobjs {
	return &Waitobjs{
		Wbs: map[int64]*WaitObject{},
	}
}

// add a object for wait return
func (ws *Waitobjs) AddWaiter(msgid int64) *WaitObject {
	ws.mtx.Lock()
	defer ws.mtx.Unlock()

	ob := WaitObject{}
	ob.wg.Add(1)
	ob.data = nil
	ws.Wbs[msgid] = &ob
	return &ob
}

// set return data
func (ws *Waitobjs) Update(msgid int64, data interface{}) {
	if ob, ok := ws.Wbs[msgid]; ok {
		ob.mtx.Lock()
		defer ob.mtx.Unlock()

		ob.data = data
		ob.wg.Done()

	}
}

// return data
func (ws *Waitobjs) GetData(msgid int64) interface{} {
	ws.mtx.Lock()
	defer ws.mtx.Unlock()

	if ob, ok := ws.Wbs[msgid]; ok {

		retdata := ob.data
		delete(ws.Wbs, msgid)
		return retdata
	}
	return nil
	//return rd.data
}

// start waiting
func (ws *Waitobjs) Waitforever(msgid int64) {
	if rd, ok := ws.Wbs[msgid]; ok {
		rd.wg.Wait()
	}
}

// start waiting
func (ws *Waitobjs) WaitTo(msgid int64, timeout time.Duration) {
	if rd, ok := ws.Wbs[msgid]; ok {
		go rd.startTimer(timeout)
		rd.wg.Wait()
	}
}

// sync mode , save return data
type WaitObject struct {
	//code   int            //ret code 0-success 1-timeout  -1-default
	data interface{}    //result
	wg   sync.WaitGroup //sync mode,flag
	mtx  sync.Mutex
}

// set return data
func (rd *WaitObject) Update(data interface{}) {
	rd.mtx.Lock()
	defer rd.mtx.Unlock()

	rd.data = data
	rd.wg.Done()

}

// return data
func (rd *WaitObject) Get() interface{} {
	return rd.data
}

// start waiting
func (rd *WaitObject) Waitforever() {
	rd.wg.Add(1)
	rd.data = nil
	rd.wg.Wait()
}

// start waiting
func (rd *WaitObject) WaitTo(timeout time.Duration) {
	rd.wg.Add(1)
	rd.data = nil
	go rd.startTimer(timeout)

	rd.wg.Wait()
}

// start  a timer for async
func (rd *WaitObject) startTimer(timeout time.Duration) {
	reqTimer := time.NewTimer(timeout)
	for {
		select {
		case <-reqTimer.C:
			rd.timeout()
		}
	}
}

// timeout result
func (rd *WaitObject) timeout() {
	rd.mtx.Lock()
	defer rd.mtx.Unlock()

	rd.data = nil
	rd.wg.Done()

}
