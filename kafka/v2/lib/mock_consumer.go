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

type mockConsumer struct {
	msgChan   chan *sarama.ConsumerMessage
	notifChan chan *cluster.Notification
	errChan   chan error
}

func createMockConsumer(brokers []string, groupID string, topics []string, config *cluster.Config) Consumer {
	return &mockConsumer{
		msgChan:   make(chan *sarama.ConsumerMessage, 100),
		notifChan: make(chan *cluster.Notification, 1),
		errChan:   make(chan error, 1),
	}
}

func (consumer *mockConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return consumer.msgChan
}

func (consumer *mockConsumer) Notifications() <-chan *cluster.Notification {
	return consumer.notifChan
}

func (consumer *mockConsumer) Errors() <-chan error {
	return consumer.errChan
}

func (consumer *mockConsumer) MarkOffset(*sarama.ConsumerMessage, string) {

}

func (consumer *mockConsumer) Close() error {
	return nil
}
