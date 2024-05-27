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

type defaultStreamBuffer struct {
	name      string
	buf       chan interface{}
	listeners []streamListener
}

func newDefaultStreamBuffer(name string, length int) *defaultStreamBuffer {
	if length < 1 {
		panic("length of defaultStreamBuffer should >= 1")
	}

	return &defaultStreamBuffer{
		name: name,
		buf:  make(chan interface{}, length),
	}
}

func (dsb *defaultStreamBuffer) Add(newItem interface{}) {
	dsb.buf <- newItem
	promProduceCounter.With(prometheus.Labels{"channel": dsb.name}).Inc()
}

func (dsb *defaultStreamBuffer) RegisterListener(l streamListener) {
	dsb.listeners = append(dsb.listeners, l)
}

func (dsb *defaultStreamBuffer) Serve() {
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

type staticStreamBuffer struct {
	name      string
	buf       chan Comparable
	state     Comparable
	listeners []streamListener
}

func newStaticStreamBuffer(name string) *staticStreamBuffer {
	return &staticStreamBuffer{
		name: name,
		buf:  make(chan Comparable),
	}
}

func (ssb *staticStreamBuffer) Add(newItem interface{}) {
	c, ok := newItem.(Comparable)
	if !ok {
		panic("non comparable object got")
	}

	ssb.buf <- c
	promProduceCounter.With(prometheus.Labels{"channel": ssb.name}).Inc()
}

func (ssb *staticStreamBuffer) RegisterListener(l streamListener) {
	ssb.listeners = append(ssb.listeners, l)
}

func (ssb *staticStreamBuffer) Serve() {
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
