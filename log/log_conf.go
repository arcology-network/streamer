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
	"fmt"

	"github.com/BurntSushi/toml"
)

type Console struct {
	SystemOut bool
}

type LocalFile struct {
	SaveFile   bool
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Ignored    string
}

var Conf *Config

// Config .
type Config struct {
	Level     string
	Console   *Console
	LocalFile *LocalFile
	Version   string
}

func InitCfg(confPath string) {
	_, err := toml.DecodeFile(confPath, &Conf)
	if err != nil {
		fmt.Printf("err=%v\n", err)
	}
	fmt.Printf("conf=%v,%v\n", Conf.Console, Conf.LocalFile)
}
