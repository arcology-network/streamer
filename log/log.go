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
	"os"
	"path"
	"path/filepath"
	"strings"

	tmos "github.com/arcology-network/consensus-engine/libs/os"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	HomeFlag = "home"
)

func InitLog(logname, logcfg, svcname, nodename string, nodeid int) {
	rootDir := viper.GetString(HomeFlag)
	//create logger
	if err := tmos.EnsureDir(path.Join(rootDir, "log"), 0777); err != nil {
		panic(err.Error())
	}
	logfile, err := os.OpenFile(path.Join(rootDir, "log", logname), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err.Error())
	}
	InitLogSystem(logfile.Name(), logcfg, svcname, nodeid, nodename)
}

func CompleteMetaInfo(svcname string) {
	err := Metas.MetaInfoToFile(svcname, GetCurrentDirectory())
	if err != nil {
		Logger.Log.Error("MetaInfoToFile err", zap.String("err", err.Error()))
	} else {
		Logger.Log.Info("MetaInfoToFile create success")
	}

}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("GetCurrentDirectory err=%v\n", err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
