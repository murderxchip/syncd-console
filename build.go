package main

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type BuildsInfo struct {
	ProjectId   string `json:"project_id"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
	SpaceId     string `json:"space_id"`
	ProjectName string `json:"project_name"`
	SpaceName   string `json:"space_name"`
}

func NewBuilds(projectName, comment, tag string) (*BuildsInfo, error) {
	var id string
	var curProject ProjectInfo
	buildInfo := &BuildsInfo{
		"",
		tag,
		comment,
		"",
		"",
		"",
	}

	project, count := NewProjects(request.Projects()).GetProjectByFuzzy(projectName)
	if count == 1 {
		curProject = project.data[0]
	} else {
		for k, v := range project.data {
			fmt.Printf("%v-项目名称:%s-空间:%s\n", k, v.ProjectName, v.SpaceName)
		}
		fmt.Printf("\033[31m 请输入具体发布的项目ID: \033[0m")
		_, err := fmt.Scanln(&id)
		if err != nil {
			return buildInfo, err
		}
		key, _ := strconv.Atoi(id)
		curProject = *project.GetProject(project.data[key].ProjectName)
	}
	buildInfo.ProjectId = strconv.Itoa(curProject.ProjectId)
	buildInfo.SpaceId = strconv.Itoa(curProject.SpaceId)
	buildInfo.ProjectName = curProject.ProjectName
	buildInfo.SpaceName = curProject.SpaceName
	return buildInfo, nil
}

func (b *BuildsInfo) getSubmitId(accessCfg AccessConfig) (taskId int) {
	//读取任务列表，找到任务id
	respData := request.ApplyList(0, 20)
	list := respData["list"]
	for _, v := range list.([]interface{}) {
		userName := v.(map[string]interface{})["username"].(string)
		projectId := int(v.(map[string]interface{})["project_id"].(float64))
		id := int(v.(map[string]interface{})["id"].(float64)) //任务id
		status := int(v.(map[string]interface{})["status"].(float64))

		if userName == accessCfg.Username && strconv.Itoa(projectId) == b.ProjectId && status == TASK_STATUS_WAIT {
			return id
		}
	}
	return 0
}

func (b *BuildsInfo) QuickBuild(accessCfg AccessConfig) error {
	request := NewRequest(accessCfg)
	err := request.SubmitById(b.ProjectId, b.SpaceId, b.Description, b.Tag)
	if err != nil {
		return err
	}

	taskId := b.getSubmitId(accessCfg)
	if taskId == 0 {
		return errors.New("未找到任务")
	}

	var build, deploy chan string
	build = make(chan string)
	deploy = make(chan string)

	defer func() {
		close(build)
		close(deploy)
	}()

	//build
	go func(taskId int) {
		print("开始构建")
		err := request.BuildStart(taskId)
		if err != nil {
			build <- "构建启动失败" + err.Error()
			return
		}
		timeCh := time.After(BUILDTIMEOUT)
		for {
			select {
			case <-timeCh:
				build <- "构建超时，请重试"
				return
			default:
				status := request.BuildStatus(taskId)
				switch status {
				case BUILD_STATUS_ERROR:
					build <- "构建出错"
					return
				case BUILD_STATUS_DONE:
					build <- "构建完成"
					return
				case BUILD_STATUS_RUNNING:
					fmt.Print(".")
				}
			}
			time.Sleep(2 * time.Second)
		}
	}(taskId)
	buildStr := <-build
	fmt.Println(buildStr)
	//构建结束，开始部署
	go func(taskId int) {
		print("开始部署")
		err := request.DeployStart(taskId)
		if err != nil {
			deploy <- "部署启动失败"
			return
		}
		timeCh := time.After(SUBMITTIMEOUT)
		for {
			select {
			case <-timeCh:
				deploy <- "部署超时，请重试"
				return
			default:
				status := request.DeployStatus(taskId)
				switch status {
				case DEPLOY_STATUS_DONE:
					deploy <- "部署成功"
					return
				case DEPLOY_STATUS_FAIL:
					deploy <- "部署失败"
					return
				case DEPLOY_STATUS_RUNNING:
					print(".")
				}
			}
			time.Sleep(2 * time.Second)
		}
	}(taskId)
	deployStr := <-deploy
	fmt.Println(deployStr)
	//fmt.Printf("\033[33m任务地址：%s\033[0m\n", DEPLOYURL+strconv.Itoa(taskId))
	return nil
}
