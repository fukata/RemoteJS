package appconfig

import (
	"os"
	"io/ioutil"
	"json"
	"fmt"
)

// 設定ファイルの値を表現する構造体
type AppConfig struct {
	ProfileName string
	MaxVirtualDesktop int
	ParallelExecCount int
	MaxRetryCount int
	FirefoxBin string
	DbHost string
	DbName string
}

// 設定ファイルを読み込む
func Parse(filename string) (AppConfig, os.Error) {
	var c AppConfig
	jsonString, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("error" + err.String())
		return c, err
	}
	err = json.Unmarshal(jsonString, &c)
	if err != nil {
		fmt.Println("error" + err.String())
		return c, err
	}
	return c, nil
}
