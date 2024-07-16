package log

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fr-str/httpea/internal/config"
)

var logFormatStr = "%s %s \x1b[90m%s\x1b[0m %s%s\n"
var timeFormat = "15:05:05.000"

const (
	TraceText = "\x1b[37;1mTRACE\x1b[0m"
	DebugText = "\x1b[35;1mDEBUG\x1b[0m"
	InfoText  = "\x1b[34;1mINFO\x1b[0m"
	WarnText  = "\x1b[33;1mWARN\x1b[0m"
	ErrorText = "\x1b[31;1mERROR\x1b[0m"
	FatalText = "\x1b[30;41;1mFATAL\x1b[0m"
)

func Debug(msg string, m ...any) {
	meta := ""
	for _, v := range m {
		meta += fmt.Sprintf(" %v", v)
	}
	if len(meta) > 2 {
		meta = meta[:len(meta)-1]
	}
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprintf(config.DebugLogFile, logFormatStr,
		time.Now().Format(timeFormat),
		DebugText,
		fmt.Sprintf("%s:%d", filepath.Base(file), line),
		msg,
		meta,
	)
}
