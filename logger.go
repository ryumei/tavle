package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/logutils"
)

type LineOfLog struct {
	RemoteAddr  string
	ContentType string
	Path        string
	Query       string
	Method      string
	Body        string
}

var TemplateOfLog = `[INFO]
Remote address:   {{.RemoteAddr}}
Content-Type:     {{.ContentType}}
HTTP method:      {{.Method}}

path:
{{.Path}}

query string:
{{.Query}}

body:             
{{.Body}}

`

// LogRequest logging HTTP request
func LogRequest(r *http.Request) {
	bufbody := new(bytes.Buffer)
	bufbody.ReadFrom(r.Body)
	body := bufbody.String()

	line := LineOfLog{
		r.RemoteAddr,
		r.Header.Get("Content-Type"),
		r.URL.Path,
		r.URL.RawQuery,
		r.Method,
		body,
	}
	tmpl, err := template.New("line").Parse(TemplateOfLog)
	if err != nil {
		panic(err)
	}

	bufline := new(bytes.Buffer)
	err = tmpl.Execute(bufline, line)
	if err != nil {
		panic(err)
	}
	log.Printf(bufline.String())
}

func openLogFile(logPath string) *os.File {
	logWriter, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return os.Stderr
	}
	return logWriter
}

// LogConfig is configuration for logging
type LogConfig struct {
	accessLog string
	serverLog string
	Level     string
}

func ConfigLogging(conf LogConfig) {
	logWriter := openLogFile(conf.serverLog)

	// Logging with logutils
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(conf.Level),
		Writer:   logWriter,
	}
	log.SetOutput(filter)
	logFlags := log.LstdFlags | log.Lmicroseconds | log.LUTC
	if filter.MinLevel == "DEBUG" {
		logFlags |= log.Lshortfile
	}
	log.SetFlags(logFlags)
}
