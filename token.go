package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

var _token string

func TokenFail() {
	RemoveToken()
	panic(fmt.Sprintf("登录失败, 请设置正确的账号密码并执行登录"))
}

func RemoveToken() {
	if err := os.Remove(TokenFile); err != nil {
		logger.Println("remove .token failed")
	}
}

func SetToken(token string) {
	//logger.Println("set token:", token)
	err := ioutil.WriteFile(TokenFile, []byte(token), 0644)
	if err != nil {
		panic(err)
	}
	_token = token
}

func GetToken() string {
	if _token == "" {
		tokenByte, err := ioutil.ReadFile(TokenFile)
		if err != nil {
			panic("请先登录")
			//NewRequest(syncdCfg.access).Login()
			//return ""
		}

		_token = string(tokenByte)
	}
	//logger.Println("get token:", _token)
	return _token
}
