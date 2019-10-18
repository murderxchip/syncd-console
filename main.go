package main

import (
	"flag"
	"fmt"
	z "github.com/nutzam/zgo"
	"strconv"
	"time"
)

const version = "1.0.0"

func help() {
	fmt.Printf(`********----------syncd外挂使用说明----------*********

  当前版本 %s

  显示当前帮助
  ./syncd-console help

  登录
  ./syncd-console login

  一键部署
  ./syncd-console submit -p test-admin-server -m "some description"

  一键部署（带标签）
  ./syncd-console submit -p prod-admin-server -m "some description" -t v2019100101
  
  查看项目列表
  ./syncd-console projects
  
  查看部署任务列表
  ./syncd-console tasks`, version)
}

func Recover() {
	if r := recover(); r != nil {
		println("\n------------------------\n执行错误,原因:")
		fmt.Printf("%s\n", r)
		println("------------------------")
		//fmt.Println(string(debug.Stack()))
	}
}

func ParseSubmitFlag(flags []string) (map[string]string, error) {
	ret := make(map[string]string)
	var k string
	flagArr := []string{"-p", "-m", "-t"}
	for _, f := range flags {
		if z.IndexOfStrings(flagArr, f) != -1 {
			k = f
		} else {
			ret[k] = f
		}
	}

	return ret, nil
}

/*
app submit -p test-admin-server -m ''
app projects
app tasks
*/
func main() {
	defer Recover()
	accessCfg := InitConfig()

	flag.Parse()

	var cmd string
	if flag.NArg() == 0 {
		cmd = "help"
	} else {
		cmd = flag.Arg(0)
	}

	switch cmd {
	case "login":
		request := NewRequest(accessCfg)
		request.Login()
		println("登录成功")
	case "submit":
		//检查参数 -p -m -t
		params, err := ParseSubmitFlag(flag.Args()[1:])
		if err != nil {
			panic(err)
		}

		if params["-p"] == "" || params["-m"] == "" {
			panic("参数错误,请输入 -p project_name -m description")
		}

		var branchName = ""
		if params["-t"] != "" {
			branchName = params["-t"]
		}

		request := NewRequest(accessCfg)
		err = request.Submit(params["-p"], params["-m"], params["-m"], branchName)
		if err != nil {
			panic("任务提交失败")
		}

		time.Sleep(time.Second * 1)

		//读取任务列表，找到任务id
		respData := request.ApplyList(0, 5)
		list := respData["list"]
		var taskId int
		for _, v := range list.([]interface{}) {
			username := v.(map[string]interface{})["username"].(string)
			projectname := v.(map[string]interface{})["project_name"].(string)
			id := int(v.(map[string]interface{})["id"].(float64)) //任务id
			status := int(v.(map[string]interface{})["status"].(float64))

			if username == accessCfg.Username && projectname == params["-p"] && status == TASK_STATUS_WAIT {
				taskId = id
				break;
			}
		}

		if taskId == 0 {
			panic("未找到任务")
		}

		var build, deploy chan int
		build = make(chan int)
		deploy = make(chan int)

		defer func() {
			close(build)
			close(deploy)
		}()

		//build
		go func(taskId int) {
			print("开始构建")
			err := request.BuildStart(taskId)
			if err != nil {
				panic("构建启动失败:" + err.Error())
			}

			for {
				select {
				case <-time.After(time.Second * 30):
					panic("构建超时，请重试")
				default:
					status := request.BuildStatus(taskId)
					switch status {
					case BUILD_STATUS_ERROR:
						panic("构建出错")
					case BUILD_STATUS_DONE:
						build <- 1
					case BUILD_STATUS_RUNNING:
						fmt.Print(".")
					}

					time.Sleep(time.Second * 2)
				}
			}
		}(taskId)
		<-build
		println("构建成功！")
		//构建结束，开始部署

		go func(taskId int) {
			print("开始部署")
			err := request.DeployStart(taskId)
			if err != nil {
				panic("部署启动失败")
			}

			for {
				select {
				case <-time.After(time.Second * 30):
					panic("部署超时，请重试")
				default:
					status := request.DeployStatus(taskId)
					switch status {
					case DEPLOY_STATUS_DONE:
						deploy <- 1
					case DEPLOY_STATUS_FAIL:
						panic("部署失败")
					case DEPLOY_STATUS_RUNNING:
						print(".")
					}

					time.Sleep(time.Second * 2)
				}
			}
		}(taskId)
		<-deploy
		println("部署成功！")

	case "projects":
		request := NewRequest(accessCfg)
		projectJson := request.Projects()
		projects := NewProjects(projectJson)
		fmt.Printf("%s - %s\n", z.AlignLeft("Project Name", 40, ' '), "Space Name")
		for _, v := range projects.data {
			fmt.Printf("%s - %s\n", z.AlignLeft(v.ProjectName, 40, ' '), v.SpaceName)
		}
	case "tasks":
		request := NewRequest(accessCfg)
		respData := request.ApplyList(0, 10)
		list := respData["list"]
		fmt.Println(z.AlignLeft("ID", 10, ' '), z.AlignLeft("Project Name", 40, ' '), z.AlignLeft("User", 30, ' '), z.AlignLeft("Submit Time", 30, ' '), "Status")
		for _, v := range list.([]interface{}) {
			username := v.(map[string]interface{})["username"].(string)
			projectname := v.(map[string]interface{})["project_name"].(string)
			id := int(v.(map[string]interface{})["id"].(float64))
			status := int(v.(map[string]interface{})["status"].(float64))
			ctime := int64(v.(map[string]interface{})["ctime"].(float64))
			t := time.Unix(ctime, 0)
			createtime := t.Format("2006-01-02 15:04:05")

			fmt.Println(z.AlignLeft(strconv.Itoa(id), 10, ' '), z.AlignLeft(projectname, 40, ' '), z.AlignLeft(username, 30, ' '), z.AlignLeft(createtime, 30, ' '), GetTaskStatusText(status))
		}
	case "help":
		help()
	default:
		help()
	}
}
