package main

import (
	"encoding/json"
	"fmt"
	"github.com/murderxchip/cmap"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const agent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:68.0) Gecko/20100101 Firefox/68.0"

const (
	BUILD_STATUS_INIT    = 0
	BUILD_STATUS_RUNNING = 1
	BUILD_STATUS_DONE    = 2
	BUILD_STATUS_ERROR   = 3

	DEPLOY_STATUS_RUNNING = 2
	DEPLOY_STATUS_DONE    = 3
	DEPLOY_STATUS_FAIL    = 4

	TASK_STATUS_WAIT    = 1
	TASK_STATUS_SUCCESS = 3
	TASK_STATUS_FAIL    = 4
	TASK_STATUS_ABANDON = 5
)

func GetTaskStatusText(status int) string {
	var statusText string
	switch status {
	case TASK_STATUS_WAIT:
		statusText = "待上线"
	case TASK_STATUS_FAIL:
		statusText = "失败"
	case TASK_STATUS_SUCCESS:
		statusText = "成功"
	case TASK_STATUS_ABANDON:
		statusText = "废弃"
	}

	return statusText
}

type RespData map[string]interface{}

type Response struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    RespData `json:"data"`
}

type ResponseArray struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}

type Request struct {
	config AccessConfig
}

type ApplyList struct {
	Username    string    `json:"username"`
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	ProjectName string    `json:"project_name"`
	Ctime       time.Time `json:"ctime"`
	Status      int       `json:"status"`
}

type BuildStatus struct {
	Status int
	Errmsg string
}

var request *Request

func NewRequest(config AccessConfig) *Request {
	if request != nil {
		return request
	}
	request = &Request{config: config}
	return request
}

func (req *Request) getUrl(route string, params cmap.CMap) string {
	if err := params.Set("_t", strconv.Itoa(int(time.Now().Unix()))); err != nil {
		panic(err)
	}

	param := make([]string, 1)
	for mapItem := range params.Dump() {
		if mapItem.Key != "" {
			param = append(param, fmt.Sprintf("%s=%s", mapItem.Key, mapItem.Value.(string)))
		}
	}

	url := fmt.Sprintf("%s://%s/%s?%s", req.config.Schema, req.config.Host, route, strings.Join(param, "&")[1:])
	//logger.Println("url:", url)
	return url
}

func ParseResponseDataArray(respBody string) ([]interface{}, error) {
	response := ResponseArray{}
	err := json.Unmarshal([]byte(respBody), &response)
	if err != nil {
		panic(err)
	}

	if response.Code == 1005 {
		TokenFail()
	}

	if response.Code != 0 {
		return nil, errors.New(response.Message)
	}

	return response.Data, nil
}

func ParseResponse(respBody string) (RespData, error) {
	response := Response{}
	err := json.Unmarshal([]byte(respBody), &response)
	if err != nil {
		panic(err)
	}

	if response.Code == 1005 {
		TokenFail()
	}

	if response.Code != 0 {
		return nil, errors.New(response.Message)
	}

	return response.Data, nil
}

func (req *Request) Login() {
	form := &LoginForm{req.config.Username, Md5(req.config.Password)}
	params := *cmap.NewCMap()
	url := req.getUrl("api/login", params)
	_, _, errs := gorequest.New().
		Post(url).
		Type("form").
		AppendHeader("Accept", "application/json").
		AppendHeader("User-Agent", agent).
		Send(fmt.Sprintf("username=%s&password=%s", form.Username, form.Password)).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}

			respData, err := ParseResponse(body)
			if err != nil {
				panic(err)
			}

			//respData
			SetToken(respData["token"].(string))
		})

	if errs != nil {
		panic("登录失败，请设置正确的用户名和密码。")
	}
}

func (req *Request) AuthCookie() *http.Cookie {
	cookie := http.Cookie{}
	cookie.Name = "_syd_identity"
	cookie.Value = GetToken()

	return &cookie
}

/**
/api/deploy/apply/project/all?_t=1568861966520
*/
func (req *Request) Projects() (projectsJson string) {
	params := *cmap.NewCMap()
	url := req.getUrl("api/deploy/apply/project/all", params)
	_, body, errs := gorequest.New().
		Get(url).
		AppendHeader("Accept", "application/json").
		AppendHeader("Host", req.config.Host).
		AppendHeader("User-Agent", agent).
		AddCookie(req.AuthCookie()).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}
		})

	if errs != nil {
		panic(errs)
	}

	respData, err := ParseResponseDataArray(body)
	if err != nil {
		//return "", err
		panic("parse remote projects failed")
	}

	//respData
	//logger.Printf("%v", respData)
	projectsByte, err := json.Marshal(respData)
	return string(projectsByte)
}

/*

 */
func (req *Request) Submit(projectName string, name string, description string, branchName string) error {
	if name == "" {
		panic("name 不能都为空")
	}
	if description == "" {
		description = name
	}
	project := NewProjects(req.Projects()).GetProject(projectName)
	if project == nil {
		panic("项目不存在")
	}

	params := *cmap.NewCMap()
	_ = params.Set("project_id", strconv.Itoa(project.ProjectId))
	_ = params.Set("space_id", strconv.Itoa(project.SpaceId))
	_ = params.Set("name", name)
	_ = params.Set("description", description)
	if branchName != "" {
		_ = params.Set("branch_name", branchName)
	}

	//fmt.Printf("%v", params)
	url := req.getUrl("api/deploy/apply/submit", params)
	_, body, errs := gorequest.New().
		Post(url).
		Type("form").
		AppendHeader("Accept", "application/json").
		AppendHeader("Host", req.config.Host).
		AppendHeader("User-Agent", agent).
		AddCookie(req.AuthCookie()).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}
		})

	if errs != nil {
		panic(errs)
	}

	_, err := ParseResponse(body)
	if err != nil {
		return err
	}

	return nil
}

/*
发布单列表 /api/deploy/apply/list?keyword=&offset=0&limit=7&_t=1569820088091
*/
func (req *Request) ApplyList(offset int, limit int) RespData {
	params := *cmap.NewCMap()
	_ = params.Set("offset", strconv.Itoa(offset))
	_ = params.Set("limit", strconv.Itoa(limit))

	url := req.getUrl("api/deploy/apply/list", params)
	//logger.Println(url)
	_, body, errs := gorequest.New().
		Get(url).
		AppendHeader("Accept", "application/json").
		AppendHeader("Host", req.config.Host).
		AppendHeader("User-Agent", agent).
		AddCookie(req.AuthCookie()).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}
		})

	if errs != nil {
		panic(errs)
	}

	respData, err := ParseResponse(body)
	if err != nil {
		panic("parse remote projects failed" + err.Error())
	}

	return respData
	//bytes, err := json.Marshal(respData["list"])
	//return string(bytes)
}

func (req *Request) BuildStart(id int) error {
	params := *cmap.NewCMap()
	url := req.getUrl("api/deploy/build/start", params)
	_, body, errs := gorequest.New().
		Post(url).
		Type("form").
		AppendHeader("Accept", "application/json").
		AppendHeader("Host", req.config.Host).
		AppendHeader("User-Agent", agent).
		AddCookie(req.AuthCookie()).
		Send(fmt.Sprintf("id=%d", id)).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}
		})

	if errs != nil {
		panic(errs)
	}

	_, err := ParseResponse(body)
	if err != nil {
		return err
	}

	return nil
}

func (req *Request) BuildStatus(id int) int {
	params := *cmap.NewCMap()
	_ = params.Set("id", strconv.Itoa(id))
	url := req.getUrl("api/deploy/build/status", params)
	_, body, errs := gorequest.New().
		Get(url).
		AppendHeader("Accept", "application/json").
		AppendHeader("Host", req.config.Host).
		AppendHeader("User-Agent", agent).
		AddCookie(req.AuthCookie()).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}
		})

	if errs != nil {
		panic(errs)
	}

	data, err := ParseResponse(body)
	if err != nil {
		panic(err)
	}

	return int(data["status"].(float64))
}

func (req *Request) DeployStart(id int) error {
	params := *cmap.NewCMap()
	url := req.getUrl("api/deploy/deploy/start", params)
	_, body, errs := gorequest.New().
		Post(url).
		Type("form").
		AppendHeader("Accept", "application/json").
		AppendHeader("Host", req.config.Host).
		AppendHeader("User-Agent", agent).
		AddCookie(req.AuthCookie()).
		Send(fmt.Sprintf("id=%d", id)).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}
		})

	if errs != nil {
		panic(errs)
	}

	_, err := ParseResponse(body)
	if err != nil {
		return err
	}

	return nil
}

func (req *Request) DeployStatus(id int) int {
	params := *cmap.NewCMap()
	_ = params.Set("id", strconv.Itoa(id))
	url := req.getUrl("api/deploy/deploy/status", params)
	_, body, errs := gorequest.New().
		Get(url).
		AppendHeader("Accept", "application/json").
		AppendHeader("Host", req.config.Host).
		AppendHeader("User-Agent", agent).
		AddCookie(req.AuthCookie()).
		End(func(response gorequest.Response, body string, errs []error) {
			if response.StatusCode != 200 {
				panic(fmt.Sprintf("%s", errs))
			}
		})

	if errs != nil {
		panic(errs)
	}

	data, err := ParseResponse(body)
	if err != nil {
		panic(err)
	}

	return int(data["status"].(float64))
}
