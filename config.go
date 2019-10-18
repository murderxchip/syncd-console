package main

import (
	"fmt"
	"github.com/astaxie/beego/config"
	z "github.com/nutzam/zgo"
	"net/url"
)

const (
	configFileName = "syncd-console.ini"
	defaultUrl     = ""
)

type AccessConfig struct {
	Schema   string
	Host     string
	Username string
	Password string
}

type SyncdConfig struct {
	configPath string
	cfg        config.Configer
	access     AccessConfig
	loaded     bool
}

var syncdCfg SyncdConfig

func InitConfig() AccessConfig {
	//println("loading config")
	syncdCfg.Load()
	return syncdCfg.access
}

func (c *SyncdConfig) Load() {
begin_load:
	cfg, err := config.NewConfig("ini", configFileName)
	if err != nil {
		//panic("配置文件加载失败")
		//init config
		accessCfg := ReadUserConfig()

		configStr := fmt.Sprintf("schema = %s\nhost = %s\nusername = %s\npassword = %s\n", accessCfg.Schema, accessCfg.Host, accessCfg.Username, accessCfg.Password)

		cfg, err = config.NewConfigData("ini", []byte(configStr))
		if err != nil {
			panic("初始化配置失败：" + err.Error())
		}

		if err = cfg.SaveConfigFile(configFileName); err != nil {
			panic("写入配置文件失败：" + err.Error())
		} else {
			println("写入配置成功！请按帮助说明进行操作。")
			goto begin_load
		}
	}

	syncdCfg.cfg = cfg

	syncdCfg.access.Schema = z.Trim(cfg.String("schema"))
	syncdCfg.access.Host = z.Trim(cfg.String("host"))
	syncdCfg.access.Username = z.Trim(cfg.String("username"))
	syncdCfg.access.Password = z.Trim(cfg.String("password"))
	if z.IsBlank(syncdCfg.access.Username) || z.IsBlank(syncdCfg.access.Host) || z.IsBlank(syncdCfg.access.Username) || z.IsBlank(syncdCfg.access.Password) {
		panic("请先设置配置文件 syncd-console.ini 的参数")
	}
}

func (c *SyncdConfig) Save() {
}

func ReadUserConfig() AccessConfig {
	var deployUrl, schema, host, username, password string
input_host:
	fmt.Printf("请输入部署主机地址（回车默认:%s）:", defaultUrl)
	if _, err := fmt.Scanln(&deployUrl); err != nil {
		deployUrl = defaultUrl
	}

	//if z.Trim(deployUrl) == "" {
	//	deployUrl = defaultUrl
	//}

	u, err := url.Parse(deployUrl)
	if err != nil || u.Scheme == "" || u.Host == "" {
		println("解析错误请重试,", err.Error())
		goto input_host
	}

	schema = u.Scheme
	host = u.Host

input_username:
	fmt.Print("请输入登录用户名:")
	if _, err := fmt.Scanln(&username); err != nil || z.Trim(username) == "" {
		goto input_username
	}

input_password:
	fmt.Print("请输入登录密码:")
	if _, err := fmt.Scanln(&password); err != nil || z.Trim(password) == "" {
		goto input_password
	}

	return AccessConfig{
		Schema:   schema,
		Host:     host,
		Username: username,
		Password: password,
	}
}
