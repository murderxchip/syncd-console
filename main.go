package main

import (
	"fmt"
	"strings"
)

const version = "1.2.0"

func help() {
	fmt.Println(`
  直接发布请输入(projectName支持模糊查询)
  d(deploy)  - 发布项目请按照格式输入[d projectName comment tag(d 项目名称 发布描述 发布分支/tag)]
  ?          - 显示当前帮助
  l(login)   - 登录
  t(task)    - 查看部署任务列表
  h(history) - 查看近期发布详情
  q(quit)    - 退出`)
}

func Recover() {
	if r := recover(); r != nil {
		println("\n------------------------\n执行错误,原因:")
		fmt.Printf("%s\n", r)
		println("------------------------")
		//fmt.Println(string(debug.Stack()))
	}
}

/*
app submit -p test-admin-server -m ''
app projects
app tasks
*/
func main() {
	var projectName, comment, tag, y string
	defer Recover()

	accessCfg,err := LoadFileInfo()
	if err != nil{
		accessCfg := ReadUserConfig()
		request := NewRequest(accessCfg)
		err := request.Login()
		if err != nil{
			panic(err.Error())
		} else {
			accessCfg.InputFile()
		}
	}else{
		request := NewRequest(accessCfg)
		err := request.Login()
		if err != nil{
			panic(err.Error())
		}
	}

	help()
	for {
		fmt.Printf("\033[35m请输入命令: \033[0m")
		count, _ := fmt.Scanln(&projectName, &comment, &tag)
		comment = strings.Replace(comment, `"`, "", -1)
		comment = strings.Replace(comment, `'`, "", -1)

		if projectName == "d" && count ==4 {
			build, err := NewBuilds(projectName, comment, tag)
			if err != nil {
				fmt.Print("\033[31m 输入错误 \033[0m\n")
				continue
			}
			fmt.Printf("空间:%s\n"+
				"项目名称:%s\n"+
				"发布描述:%s\n"+
				"tag:%s\n"+
				"\033[31m确定发布？(y/n) : \033[0m",
				build.SpaceName, build.ProjectName, build.Description, build.Tag)
			_, err = fmt.Scanln(&y)
			if y == "y" || y == "Y" {
				err = build.QuickBuild(accessCfg)
				if err != nil {
					fmt.Print("\033[31m 发布失败 \033[0m\n")
				}
			}

		} else {
			switch projectName {
			case "q", "Q", "quit":
				Quit()
				return
			case "l","login":
				request.Login()
			case "t","task":
				ShowProjectList()
			case "h","history":
				ShowProjectInfo()
			default:
				help()
			}
		}
	}
}
