package main

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "syncd-console", log.Ltime|log.Ldate)
}
