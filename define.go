package main

import (
	"os"
	"time"
)

//构建超时时间
var BUILDTIMEOUT = 2 * time.Minute

//提交超时时间
var SUBMITTIMEOUT = 2 * time.Minute

var HOME = os.Getenv("HOME") + "/"

var (
	ConfigFileName = HOME + ".syncd-console.ini"
	DefaultUrl     = "http://your-syncd-host/entry"
	TokenFile      = ".syncd-token"
)
