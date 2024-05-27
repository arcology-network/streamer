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
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

var asyncProducerCreator func([]string, *sarama.Config) (sarama.AsyncProducer, error)
var consumerCreator func([]string, string, []string, *cluster.Config) (Consumer, error)

func init() {
	asyncProducerCreator = func(brokers []string, config *sarama.Config) (sarama.AsyncProducer, error) {
		return sarama.NewAsyncProducer(brokers, config)
	}
	consumerCreator = func(brokers []string, groupID string, topics []string, config *cluster.Config) (Consumer, error) {
		return cluster.NewConsumer(brokers, groupID, topics, config)
	}
}
