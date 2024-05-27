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

package log

import (
	"io/ioutil"
	"sync"
)

var Metas *MetaInfos

type MetaInfoItem struct {
	ReceptivMsgs    []string
	DeliverableMsgs []string
}

type MetaInfos struct {
	Mis         map[string]*MetaInfoItem
	WorkThreads []string
	lock        sync.RWMutex
}

func InitMetaInfos() {
	Metas = &MetaInfos{
		Mis: map[string]*MetaInfoItem{},
	}
}

func (mi *MetaInfos) Add(workthread, msgname string, isSend bool) {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	mii, ok := mi.Mis[workthread]
	if !ok {
		mii = &MetaInfoItem{
			ReceptivMsgs:    []string{},
			DeliverableMsgs: []string{},
		}
		mi.Mis[workthread] = mii
		mi.WorkThreads = append(mi.WorkThreads, workthread)
	}
	if isSend {
		appendString(&mii.DeliverableMsgs, msgname)
	} else {
		appendString(&mii.ReceptivMsgs, msgname)
	}

}

func (mi *MetaInfos) MetaInfoToFile(servicename, configfilePath string) error {
	mi.lock.Lock()
	defer mi.lock.Unlock()

	outputString := ""
	for _, workthreadname := range mi.WorkThreads {
		if metainfoItem, ok := mi.Mis[workthreadname]; ok {
			for _, msgname := range metainfoItem.DeliverableMsgs {
				outputString += servicename + "\t" + workthreadname + "\t" + msgname + "\tsend" + "\n"
			}
			for _, msgname := range metainfoItem.ReceptivMsgs {
				outputString += servicename + "\t" + workthreadname + "\t" + msgname + "\treceive" + "\n"
			}
		}
	}

	return ioutil.WriteFile(configfilePath+"/"+servicename+".conf", []byte(outputString), 0666)

}

func appendString(list *[]string, one string) {
	for _, v := range *list {
		if v == one {
			return
		}
	}
	(*list) = append((*list), one)
}
