//------------------------------------------------------------------------------
//
//  Copyright 2020 by International Games System Co., Ltd.
//  All rights reserved.
//
//  This software is the confidential and proprietary information of
//  International Game System Co., Ltd. ('Confidential Information'). You shall
//  not disclose such Confidential Information and shall use it only in
//  accordance with the terms of the license agreement you entered into with
//  International Game System Co., Ltd.
//
//------------------------------------------------------------------------------

//------------------------------------------------------------------------------
//	Package declare
//------------------------------------------------------------------------------

package agency

//------------------------------------------------------------------------------
//	Import packages
//------------------------------------------------------------------------------

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shiena/ansicolor"
)

//------------------------------------------------------------------------------
//	Constants
//------------------------------------------------------------------------------

const (
	fgBoldBlack   string = "\x1b[30;1m"
	fbBoldRed     string = "\x1b[31;1m"
	fgBoldGreen   string = "\x1b[32;1m"
	fgBoldYellow  string = "\x1b[33;1m"
	fgBoldBlue    string = "\x1b[34;1m"
	fgBoldMagenta string = "\x1b[35;1m"
	fgBoldCyan    string = "\x1b[36;1m"
	fgBoldWhite   string = "\x1b[37;1m"
	fbBlack       string = "\x1b[30m"
	fgRed         string = "\x1b[31m"
	fgGreen       string = "\x1b[32m"
	fgYellow      string = "\x1b[33m"
	fgBlue        string = "\x1b[34m"
	fgMagenta     string = "\x1b[35m"
	fgCyan        string = "\x1b[36m"
	fgWhite       string = "\x1b[37m"
	fgNormal      string = "\x1b[0m"
)

const (
	_ = iota
	critical
	err
	warn
	notice
	info
	debug
)

const (
	onceFetchLines = 200
)

//------------------------------------------------------------------------------
//	Variables
//------------------------------------------------------------------------------

var (
	logLevels = map[string]int{
		"DEBUG":  6,
		"INFO":   5,
		"NOTICE": 4,
		"WARN":   3,
		"ERROR":  2,
		"CRIT":   1,
	}
	logLevel    = 6
	timeFormat  = "2006-01-02 15:04:05"
	mutex       = &sync.RWMutex{}
	outFilename = ""
	fileOpened  = false
	outFile     = os.NewFile(0, "")
	outDetail   = false
	colorable   = true
	logStrings  = NewConcurrentArray()
	colorWriter = ansicolor.NewAnsiColorWriter(os.Stdout)
	_, b, _, _  = runtime.Caller(0)
	basepath    = filepath.Dir(b) + "\\.."
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

type (
	logMessage struct {
		format string
		v      []interface{}
		file   string
		line   int
		level  int
	}
)

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// LoggerSetup : 設定 log 的一些基礎設定
// @param	hasDetail	是否印出所在檔案與行號
// @param	hasFile		是否將紀錄寫入檔案（自動）
// @param	logLevel	記錄等級
func LoggerSetup(hasDetail, hasFile bool, logLevel string) {
	outDetail = hasDetail
	if hasFile {
		OpenLogFile()
	}
	SetLogLevel(logLevel)
}

func Debug(format string, v ...interface{}) {
	if logLevel == debug {
		writeLog(createMessage(format, v, 6))
	}
}

func Info(format string, v ...interface{}) {
	if logLevel >= info {
		writeLog(createMessage(format, v, 5))
	}
}

func Notice(format string, v ...interface{}) {
	if logLevel >= notice {
		writeLog(createMessage(format, v, 4))
	}
}

func Warn(format string, v ...interface{}) {
	if logLevel >= warn {
		writeLog(createMessage(format, v, 3))
	}
}

func Error(format string, v ...interface{}) {
	if logLevel >= err {
		writeLog(createMessage(format, v, 2))
	}
}

func Critical(format string, v ...interface{}) {
	if logLevel >= critical {
		writeLog(createMessage(format, v, 1))
	}
}

func SetLogLevel(level string) error {
	level = strings.ToUpper(level)
	if numLevel, ok := logLevels[level]; ok {
		logLevel = numLevel
		return nil
	}
	return errors.New("Invalid log level: " + level)
}

func SetTimeFormat(format string) {
	timeFormat = format
}

func OpenLogFile() {
	createLogFile()
}

func CloseLogFile() {
	outFile.Close()
}

func GetAppName() string {
	strs := strings.Split(os.Args[0], "/")
	if len(strs) == 1 {
		strs = strings.Split(os.Args[0], "\\")
	}
	return strings.Split(strs[len(strs)-1], ".")[0]
}

func GetLogContents(index int) ([]string, int) {
	var gain []interface{}
	pos := 0
	if index == -1 {
		lines := logStrings.Len()
		if lines < onceFetchLines {
			pos = 0
		} else {
			pos = lines - onceFetchLines - 1
		}
	} else {
		pos = index
	}
	gain = logStrings.GainFrom(pos, onceFetchLines)

	length := len(gain)
	res := make([]string, length)
	for i := 0; i < length; i++ {
		res[i] = gain[i].(string)
	}
	return res, pos
}

//------------------------------------------------------------------------------
//	Private Methods
//------------------------------------------------------------------------------

func createMessage(format string, v []interface{}, level int) *logMessage {
	file, line := getFileAndLine()
	return &logMessage{format, v, file, line, level}
}

func getFileAndLine() (string, int) {
	_, file, line, _ := runtime.Caller(3)
	rel, _ := filepath.Rel(basepath, file)
	return rel, line
}

func writeLog(data *logMessage) {
	mutex.Lock()
	defer mutex.Unlock()
	cout, fout := createLogString(data)
	fmt.Fprintln(colorWriter, cout)
	if fileOpened {
		outFile.WriteString(fout + "\n")
	}
	logStrings.Append(fout)
}

func createLogString(data *logMessage) (string, string) {
	var color, level string
	switch data.level {
	case debug:
		color = fgCyan
		level = "[D]"
	case info:
		color = ""
		level = "[I]"
	case notice:
		color = fgGreen
		level = "[N]"
	case warn:
		color = fgYellow
		level = "[W]"
	case err:
		color = fgRed
		level = "[E]"
	case critical:
		color = fbBoldRed
		level = "[C]"
	}
	now := time.Now().Format(timeFormat)
	out := ""
	if outDetail {
		out = fmt.Sprint(level, " ", now, " ", data.file, ":", data.line, "  ▶  ")
	} else {
		out = fmt.Sprint(level, " ", now, "  ▶  ")
	}

	if len(data.v) > 0 {
		out = out + fmt.Sprintf(data.format, data.v...)
	} else {
		out = out + data.format
	}
	if data.level != info && colorable {
		return fmt.Sprint(color, out, fgNormal), out
	}
	return out, out
}

func createLogFile() {
	mutex.Lock()
	defer mutex.Unlock()
	if fileOpened {
		fmt.Println("log:createLogFile: already exist.")
		return
	}
	if _, err := os.Stat(outFilename); os.IsNotExist(err) {
		outFile, err = os.Create(outFilename)
		if err != nil {
			fmt.Println("log:createLogFile: cannot create log file. FILE=", outFilename)
			return
		}
		fileOpened = true
	}
	fmt.Println("log:createLogFile: file exist. FILE=", outFilename)
}

//------------------------------------------------------------------------------
//	Auto initialize function
//------------------------------------------------------------------------------

func init() {
	now := time.Now().Format("20060102-150405")
	outFilename = fmt.Sprintf("%s_%s.txt", GetAppName(), now)
}
