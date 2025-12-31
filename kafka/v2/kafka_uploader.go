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

package kafka

import (
	"strings"

	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/kafka/v2/lib"
	"github.com/spf13/viper"
)

type KafkaUploader struct {
	outgoing  *lib.ComOutgoing
	relations map[string]string
	Name      string
}

// return a Subscriber struct
func NewKafkaUploader(name string, relations map[string]string) *KafkaUploader {
	uploader := KafkaUploader{}
	uploader.relations = relations
	uploader.Name = name
	return &uploader
}

func (ku *KafkaUploader) OnStart() {
	mqaddr := viper.GetString("mqaddr")
	ku.outgoing = new(lib.ComOutgoing)
	if err := ku.outgoing.Start(strings.Split(mqaddr, ","), ku.relations, ku.Name); err != nil {
		panic(err)
	}
}

func (ku *KafkaUploader) OnMessageArrived(msgs []*actor.Message) error {
	ku.outgoing.Send(msgs[0])
	return nil
}
