package main

import "encoding/json"

type ProjectInfo struct {
	ProjectId   int    `json:"project_id"`
	ProjectName string `json:"project_name"`
	SpaceId     int    `json:"space_id"`
	SpaceName   string `json:"space_name"`
}

type Projects struct {
	data []ProjectInfo
}

func NewProjects(data string) *Projects {
	projects := &Projects{}
	err := json.Unmarshal([]byte(data), &projects.data)
	if err != nil {
		panic("project data error")
	}
	return projects
}

func (p *Projects) GetProject(name string) *ProjectInfo {
	for _, v := range p.data {
		if v.ProjectName == name {
			return &v
		}
	}
	return nil
}
