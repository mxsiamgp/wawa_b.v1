package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DEFAULT_CONFIG_FILENAME = "default.json"
)

func NewConfig(confDir, prof string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("json")

	defCfgAbsFn := filepath.Join(confDir, DEFAULT_CONFIG_FILENAME)
	if _, err := os.Stat(defCfgAbsFn); err != nil && os.IsNotExist(err) {
		panic("指定的配置文件目录中不存在默认配置文件")
	}

	profAbsFn := filepath.Join(confDir, fmt.Sprintf("%s.json", prof))
	if _, err := os.Stat(profAbsFn); err != nil && os.IsNotExist(err) {
		panic("指定的配置文件目录中不存在指定的配置文件")
	}

	defCfgFile, err := os.Open(defCfgAbsFn)
	if err != nil {
		panic(err)
	}
	defer defCfgFile.Close()

	if err := v.ReadConfig(bufio.NewReader(defCfgFile)); err != nil {
		panic(err)
	}

	profCfgFile, err := os.Open(profAbsFn)
	if err != nil {
		panic(err)
	}
	defer profCfgFile.Close()

	if err := v.MergeConfig(bufio.NewReader(profCfgFile)); err != nil {
		panic(err)
	}

	return v
}
