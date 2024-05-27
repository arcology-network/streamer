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

type DefaultProducer struct {
	name          string
	outputs       []string
	bufferLengths []int
}

func NewDefaultProducer(name string, outputs []string, bufferLengths []int) StreamProducer {
	return &DefaultProducer{
		name:          name,
		outputs:       outputs,
		bufferLengths: bufferLengths,
	}
}

func (p *DefaultProducer) GetName() string {
	return p.name
}

func (p *DefaultProducer) GetOutputs() []string {
	return p.outputs
}

func (p *DefaultProducer) GetBufferLengths() []int {
	return p.bufferLengths
}
