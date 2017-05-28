package snetlog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/logging"
)

//Logger ...
type Logger interface {
	Trace(args ...interface{})
	Tracef(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Notice(args ...interface{})
	Noticef(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Erro(args ...interface{})
	Errof(format string, args ...interface{})
	Alert(args ...interface{})
	Alertf(format string, args ...interface{})
	Emergency(args ...interface{})
	Emergencyf(format string, args ...interface{})
}

//formats ...
var formats = map[logging.Severity]string{
	logging.Debug:     "[TRACE] ",
	logging.Info:      "[INFO] ",
	logging.Notice:    "[NOTICE] ",
	logging.Warning:   "[WARN] ",
	logging.Error:     "[ERROR] ",
	logging.Alert:     "[ALERT] ",
	logging.Emergency: "[EMERGENCY] ",
}

//Log ...
type Log struct {
	enableConsole     bool
	enableFile        bool
	enableStackDriver bool

	muxConsole sync.Mutex

	muxFile    sync.Mutex
	fileName   string
	bufferFile *bytes.Buffer

	logStackdriver *logging.Logger
}

//f retorna um string apresentando t no formato DD/MM/AAAA hh:mm:ss.milisegundo
func formataTimePadraoLog(t time.Time) string {
	var momento = fmt.Sprintf("%02d/%02d/%04d %02d:%02d:%02d.%d",
		t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6)
	return momento
}

func flushLogFile(log *Log) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.muxFile.Lock()
			if log.bufferFile != nil {
				if log.bufferFile.Len() > 0 && log.fileName != "" {
					ioutil.WriteFile(log.fileName, log.bufferFile.Bytes(), os.ModeAppend)
				}
				log.bufferFile.Reset()
				log.muxFile.Unlock()
			}
		}
	}
}

func (l *Log) output(severity logging.Severity, format string, args ...interface{}) {
	if l.enableStackDriver && l.logStackdriver != nil {
		if format != "" {
			l.logStackdriver.Log(logging.Entry{Severity: severity,
				Payload: fmt.Sprintf(format, args...)})
		} else {
			if l.logStackdriver != nil {
				l.logStackdriver.Log(logging.Entry{Severity: severity, Payload: fmt.Sprint(args...)})
			}
		}
	}

	if l.enableConsole {
		l.muxConsole.Lock()
		log.Printf(formats[severity])
		fmt.Printf(formataTimePadraoLog(time.Now()) + ": ")
		if format == "" {
			fmt.Printf(fmt.Sprint(args...))
		} else {
			fmt.Printf(format, args...)
		}
		fmt.Printf("\n")
		l.muxConsole.Unlock()
	}

	if l.enableFile && l.fileName != "" && l.bufferFile != nil {
		l.muxFile.Lock()
		l.bufferFile.WriteString(formats[severity])
		l.bufferFile.WriteString(formataTimePadraoLog(time.Now()) + ": ")
		if format == "" {
			l.bufferFile.WriteString(fmt.Sprint(args...))
		} else {
			l.bufferFile.WriteString(fmt.Sprintf(format, args...))
		}
		l.bufferFile.WriteString("\n")
		l.muxFile.Unlock()
	}
}

//Trace ...
func (l *Log) Trace(args ...interface{}) {
	l.output(logging.Debug, "", args...)
}

//Tracef ...
func (l *Log) Tracef(format string, args ...interface{}) {
	l.output(logging.Debug, format, args...)
}

//Info ...
func (l *Log) Info(args ...interface{}) {
	l.output(logging.Info, "", args...)
}

//Infof ...
func (l *Log) Infof(format string, args ...interface{}) {
	l.output(logging.Debug, format, args...)
}

//Notice ...
func (l *Log) Notice(args ...interface{}) {
	l.output(logging.Notice, "", args...)
}

//Noticef ...
func (l *Log) Noticef(format string, args ...interface{}) {
	l.output(logging.Notice, format, args...)
}

//Erro ...
func (l *Log) Erro(args ...interface{}) {
	l.output(logging.Error, "", args...)
}

//Errof ...
func (l *Log) Errof(format string, args ...interface{}) {
	l.output(logging.Error, format, args...)
}

//Warn ...
func (l *Log) Warn(args ...interface{}) {
	l.output(logging.Warning, "", args...)
}

//Warnf ...
func (l *Log) Warnf(format string, args ...interface{}) {
	l.output(logging.Warning, format, args...)
}

//Alert ...
func (l *Log) Alert(args ...interface{}) {
	l.output(logging.Alert, "", args...)
}

//Alertf ...
func (l *Log) Alertf(format string, args ...interface{}) {
	l.output(logging.Alert, format, args...)
}

//Emergency ...
func (l *Log) Emergency(args ...interface{}) {
	l.output(logging.Emergency, "", args...)
}

//Emergencyf ...
func (l *Log) Emergencyf(format string, args ...interface{}) {
	l.output(logging.Emergency, format, args...)
}

//FileConfig ...
type FileConfig struct {
	FileName string
}

//NewLogFile ...
func NewLogFile(config *FileConfig) *Log {
	log := &Log{
		enableFile: true,
		fileName:   config.FileName,
		bufferFile: bytes.NewBuffer(make([]byte, 0, 3072)),
	}

	go flushLogFile(log)

	return log
}

//ConsoleConfig ...
type ConsoleConfig struct {
	FileName string
}

//NewLogConsole ...
func NewLogConsole(config *ConsoleConfig) *Log {
	log := &Log{
		enableConsole: true,
	}

	return log
}
