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

package broker

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type streamBuffer interface {
	Add(newItem interface{})
	RegisterListener(l streamListener)
	Serve()
}

type DefaultStreamBuffer struct {
	name      string
	buf       chan interface{}
	listeners []streamListener
}

func NewDefaultStreamBuffer(name string, length int) *DefaultStreamBuffer {
	if length < 1 {
		panic("length of defaultStreamBuffer should >= 1")
	}

	return &DefaultStreamBuffer{
		name: name,
		buf:  make(chan interface{}, length),
	}
}

func (dsb *DefaultStreamBuffer) Add(newItem interface{}) {
	dsb.buf <- newItem
	promProduceCounter.With(prometheus.Labels{"channel": dsb.name}).Inc()
}

func (dsb *DefaultStreamBuffer) RegisterListener(l streamListener) {
	dsb.listeners = append(dsb.listeners, l)
}

func (dsb *DefaultStreamBuffer) Serve() {
	for item := range dsb.buf {
		for _, l := range dsb.listeners {
			l.Notify(item)
		}
		promConsumeCounter.With(prometheus.Labels{"channel": dsb.name}).Inc()
	}
}

type Comparable interface {
	Equals(rhs Comparable) bool
}

type StaticStreamBuffer struct {
	name      string
	buf       chan Comparable
	state     Comparable
	listeners []streamListener
}

func NewStaticStreamBuffer(name string) *StaticStreamBuffer {
	return &StaticStreamBuffer{
		name: name,
		buf:  make(chan Comparable),
	}
}

func (ssb *StaticStreamBuffer) Add(newItem interface{}) {
	c, ok := newItem.(Comparable)
	if !ok {
		panic("non comparable object got")
	}

	ssb.buf <- c
	promProduceCounter.With(prometheus.Labels{"channel": ssb.name}).Inc()
}

func (ssb *StaticStreamBuffer) RegisterListener(l streamListener) {
	ssb.listeners = append(ssb.listeners, l)
}

func (ssb *StaticStreamBuffer) Serve() {
	for item := range ssb.buf {
		promConsumeCounter.With(prometheus.Labels{"channel": ssb.name}).Inc()
		if ssb.state != nil && ssb.state.Equals(item) {
			log.WithFields(log.Fields{
				"channel": ssb.name,
			}).Warn("duplicate state got")
			continue
		}

		ssb.state = item
		for _, l := range ssb.listeners {
			l.Notify(item)
		}
	}
}
