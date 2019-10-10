package util

import (
	"encoding/json"
	"log"
)

// PanicJson 快速打印出一个变量，并直接退出，加速调试
func PanicJson(a interface{}) {
	bs, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		panic(err)
	}
	panic(string(bs))
}

// 打印error信息
func PrintError(err error) {
	log.Println("Error: " + err.Error())
}

// 打印进度信息
func PrintProcess(info string) {
	log.Println("=== " + info + "===")
}