package main

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/howeyc/gopass"
	z "github.com/nutzam/zgo"
	"github.com/pkg/errors"
	"net/url"
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

func LoadFileInfo() (AccessConfig,error)  {
	err := syncdCfg.LoadFile()

	return syncdCfg.access,err
}

func (c *AccessConfig) InputFile()  {
	configStr := fmt.Sprintf("schema = %s\nhost = %s\nusername = %s\npassword = %s\n", c.Schema, c.Host, c.Username, c.Password)
	cfg, err := config.NewConfigData("ini", []byte(configStr))
	if err != nil {
		panic("初始化配置失败：" + err.Error())
	}
	if err = cfg.SaveConfigFile(ConfigFileName); err != nil {
		panic("写入配置文件失败：" + err.Error())
	} else {
		println("写入配置成功！请按帮助说明进行操作。")
	}
	return
}

func (c *SyncdConfig)LoadFile() error {
	cfg, err := config.NewConfig("ini", ConfigFileName)
	if err != nil {
		return errors.New("配置文件错误")
	}
	syncdCfg.cfg = cfg

	syncdCfg.access.Schema = z.Trim(cfg.String("schema"))
	syncdCfg.access.Host = z.Trim(cfg.String("host"))
	syncdCfg.access.Username = z.Trim(cfg.String("username"))
	syncdCfg.access.Password = z.Trim(cfg.String("password"))
	if z.IsBlank(syncdCfg.access.Username) || z.IsBlank(syncdCfg.access.Host) || z.IsBlank(syncdCfg.access.Username) || z.IsBlank(syncdCfg.access.Password) {
		return errors.New("请先设置配置文件 syncd-console.ini 的参数")
	}
	return nil
}

func (c *SyncdConfig) Load() {
begin_load:
	cfg, err := config.NewConfig("ini", ConfigFileName)
	if err != nil {
		//panic("配置文件加载失败")
		//init config
		accessCfg := ReadUserConfig()

		configStr := fmt.Sprintf("schema = %s\nhost = %s\nusername = %s\npassword = %s\n", accessCfg.Schema, accessCfg.Host, accessCfg.Username, accessCfg.Password)

		cfg, err = config.NewConfigData("ini", []byte(configStr))
		if err != nil {
			panic("初始化配置失败：" + err.Error())
		}

		if err = cfg.SaveConfigFile(ConfigFileName); err != nil {
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
	fmt.Printf("请输入部署主机地址（例子:%s）:", DefaultUrl)
	if _, err := fmt.Scanln(&deployUrl); err != nil {
		deployUrl = DefaultUrl
	}

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
	passwordByte, err := gopass.GetPasswdMasked()
	password = string(passwordByte)

	if err != nil || z.Trim(password) == "" {
		goto input_password
	}

	return AccessConfig{
		Schema:   schema,
		Host:     host,
		Username: username,
		Password: password,
	}
}
