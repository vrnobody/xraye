package latency

import (
	"log"
	"os"

	"github.com/xtls/xray-core/common/serial"
)

var stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)

const (
	Loglevel_None    int = 0
	Loglevel_Error   int = 1
	Loglevel_Warning int = 2
	Loglevel_Info    int = 3
	Loglevel_Debug   int = 4
)

type Logger struct {
	tag      string
	loglevel int
}

func (l *Logger) logln(tag string, msg ...any) {
	if len(l.tag) > 0 {
		stdlog.Printf("%s %s %s\n", tag, l.tag, serial.Concat(msg...))
	} else {
		stdlog.Printf("%s %s\n", tag, serial.Concat(msg...))
	}
}

func (l *Logger) Error(msg ...any) {
	if l.loglevel < Loglevel_Error {
		return
	}
	l.logln("[Error]", msg...)
}

func (l *Logger) Warn(msg ...any) {
	if l.loglevel < Loglevel_Warning {
		return
	}
	l.logln("[Warn]", msg...)
}

func (l *Logger) Info(msg ...any) {
	if l.loglevel < Loglevel_Info {
		return
	}
	l.logln("[Info]", msg...)
}

func (l *Logger) Debug(msg ...any) {
	if l.loglevel < Loglevel_Debug {
		return
	}
	l.logln("[Debug]", msg...)
}
