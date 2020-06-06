package main

import (
	"fmt"
	z "github.com/nutzam/zgo"
	"strconv"
	"time"
)

//Quit
func Quit() {
	fmt.Println("退出成功")
}

//ShowProjectList
func ShowProjectList() ([]int, []int) {
	//获取项目空间列表
	accessCfg := InitConfig()
	request := NewRequest(accessCfg)
	spaceList := request.SpaceList()
	var projectArr []int
	var spaceArr []int
	idx := 0
	for _, v := range spaceList.Data.List {
		fmt.Printf("---------%v---------\n", v.Name)
		proList := request.ProjectsList(v.ID)
		for _, project := range proList.Data.List {
			fmt.Printf("%v-%s\n", idx, project.Name)
			idx++
			projectArr = append(projectArr, project.ID)
		}
		spaceArr = append(spaceArr, v.ID)
		fmt.Println()
	}

	return projectArr, spaceArr
}

//ShowProjectInfo
func ShowProjectInfo() {
	accessCfg := InitConfig()
	request := NewRequest(accessCfg)
	respData := request.ApplyList(0, 10)
	list := respData["list"]
	fmt.Println(z.AlignLeft("ID", 10, ' '), z.AlignLeft("Project Name", 20, ' '), z.AlignLeft("User", 20, ' '), z.AlignLeft("Submit Time", 20, ' '), "Status")
	for _, v := range list.([]interface{}) {
		userName := v.(map[string]interface{})["username"].(string)
		projectName := v.(map[string]interface{})["project_name"].(string)
		id := int(v.(map[string]interface{})["id"].(float64))
		status := int(v.(map[string]interface{})["status"].(float64))
		ctime := int64(v.(map[string]interface{})["ctime"].(float64))
		t := time.Unix(ctime, 0)
		createtime := t.Format("2006-01-02 15:04:05")

		fmt.Println(z.AlignLeft(strconv.Itoa(id), 10, ' '), z.AlignLeft(projectName, 20, ' '), z.AlignLeft(userName, 20, ' '), z.AlignLeft(createtime, 20, ' '), GetTaskStatusText(status))

	}
}
