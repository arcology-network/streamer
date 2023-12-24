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
