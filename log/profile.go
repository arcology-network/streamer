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
	"runtime/pprof"
	"time"

	"github.com/shirou/gopsutil/mem"
	"github.com/spf13/viper"
)

//****************************************************************************
//flags.String("memprofile", "memprofile", "memprofile file name")
//flags.Float64("threshold", 30.0, "percent of memory used")
//flags.Int("ticker", 2, "tick per nums seconds")
//****************************************************************************

func RecordProfile() {
	ticker := time.NewTicker(time.Duration(viper.GetInt("ticker")) * time.Second)
	savePercent := float64(0.0)
	threshold := viper.GetFloat64("threshold")
	go func() {
		for {
			select {
			case <-ticker.C:
				info, _ := mem.VirtualMemory()

				if savePercent > 0.0 && info.UsedPercent-savePercent >= threshold {
					writeMemProfile()
				}
				savePercent = info.UsedPercent
			}
		}
	}()
}

func writeMemProfile() {
	Logger.AddLog(0, LogLevel_Info, "service-init", "starter", "&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&", LogType_Inlog, 0, 0, 0, 0)
	f, err := os.Create(viper.GetString("memprofile"))
	if err != nil {
		Logger.AddLog(0, LogLevel_Error, "service-init", "starter", fmt.Sprintf("could not create memory profile: %v", err), LogType_Inlog, 0, 0, 0, 0)
		return
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.WriteHeapProfile(f); err != nil {
		Logger.AddLog(0, LogLevel_Error, "service-init", "starter", "could not write memory profile:", LogType_Inlog, 0, 0, 0, 0)
	}
}
