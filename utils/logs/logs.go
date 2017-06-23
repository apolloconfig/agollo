package logs

import (
	"bytes"
	"log"
)


func CreateLogger() *log.Logger  {
	var buf bytes.Buffer
	return  log.New(&buf, "logger: ", log.Lshortfile)
}
