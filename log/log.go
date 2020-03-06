package log

import (
	"fmt"
	"github.com/kbrownehs18/gotools/common"
	"github.com/kbrownehs18/gotools/constants"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Level log level
type Level int

const (
	// TRACE log level
	TRACE Level = 1 << iota
	// DEBUG log level
	DEBUG
	// INFO log level
	INFO
	// WARNING log level
	WARNING
	// ERROR log level
	ERROR
	// FATAL log level
	FATAL
)

// NewLevel new log level
func NewLevel(name string) Level {
	name = strings.ToUpper(name)
	switch name {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	}

	return TRACE
}

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}

	return "TRACE"
}

// Rotate log
type Rotate int

const (
	// NONE rotate
	NONE Rotate = 1 << iota
	// SIZE rotate
	SIZE
	// DAILY rotate
	DAILY
)

// NewRotate new a rotate
func NewRotate(name string) Rotate {
	name = strings.ToUpper(name)
	switch name {
	case "SIZE":
		return SIZE
	case "DAILY":
		return DAILY
	}

	return NONE
}

// Appender log
type Appender int

const (
	// CONSOLE appender
	CONSOLE Appender = 1 << iota
	// FILE appender
	FILE
)

// NewAppender new appender type
func NewAppender(name string) Appender {
	name = strings.ToUpper(name)
	if name == "FILE" {
		return FILE
	}

	return CONSOLE
}

// Logger log struct
type Logger struct {
	appender    Appender
	level       Level
	logger      *log.Logger
	fileHandler *FileHandler
	name        string
}

// Level return log level
func (l *Logger) Level() Level {
	return l.level
}

// FileHandler log file
type FileHandler struct {
	fd          *os.File
	fileName    string
	rotate      Rotate
	maxBytes    int
	backupCount int
	lock        *sync.Mutex
}

func (fh *FileHandler) Write(b []byte) (n int, err error) {
	fh.rollover()
	return fh.fd.Write(b)
}

// Close file log handler close
func (fh *FileHandler) Close() error {
	if fh.fd != nil {
		return fh.fd.Close()
	}
	return nil
}

func (fh *FileHandler) rollover() {
	if fh.rotate == NONE {
		return
	}

	fh.lock.Lock()
	defer fh.lock.Unlock()

	f, err := fh.fd.Stat()
	if err != nil {
		return
	}

	if fh.rotate == SIZE {
		// rotating log by file size
		if fh.maxBytes <= 0 {
			// unlimited
			return
		} else if f.Size() < int64(fh.maxBytes) {
			// no reach max limit
			return
		}
	} else if fh.rotate == DAILY {
		if common.TimeFormat(f.ModTime(), 1) == common.Now(1) {
			// date is same
			return
		}
	}

	fh.fd.Close()
	if fh.backupCount > 0 {
		if fh.rotate == SIZE {
			for i := fh.backupCount - 1; i > 0; i-- {
				sfn := fmt.Sprintf("%s.%d", fh.fileName, i)
				dfn := fmt.Sprintf("%s.%d", fh.fileName, i+1)
				if common.Exists(sfn) {
					os.Rename(sfn, dfn)
				}
			}

			dfn := fmt.Sprintf("%s.1", fh.fileName)
			os.Rename(fh.fileName, dfn)
		} else if fh.rotate == DAILY {
			// remove
			os.Remove(fmt.Sprintf("%s.%s", fh.fileName,
				common.TimeFormat(f.ModTime().Add(0-time.Duration(fh.backupCount-1)*time.Hour*24), 1)))
			os.Rename(fh.fileName, fmt.Sprintf("%s.%s", fh.fileName,
				common.TimeFormat(time.Now().Add(0-time.Duration(1)*time.Hour*24), 1)))
		}
	} else {
		os.Remove(fh.fileName)
	}

	fh.fd, _ = os.OpenFile(fh.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}

// NewFileHandler new FileHandler
func NewFileHandler(path, fileName, rotate string, backupCount int, logSize ...int) (*FileHandler, error) {
	if !common.Exists(path) {
		if err := os.MkdirAll(path, 0777); err != nil {
			return nil, err
		}
	}

	fileName = path + constants.PathSeparator + fileName
	fd, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	size := 200 << 20 // default 200M
	if len(logSize) > 0 {
		size = logSize[0]
	}
	return &FileHandler{fd: fd, fileName: fileName, rotate: NewRotate(rotate),
		maxBytes: size, backupCount: backupCount, lock: new(sync.Mutex)}, nil
}

// NewLogger new a logger
// name log name
// appender output console or file
// level output log level
// args FileHandler
func NewLogger(name, appender, level string, args ...interface{}) (*Logger, error) {
	a := NewAppender(appender)

	var output io.Writer
	var fileHandler *FileHandler
	var err error
	if a == FILE {
		output = nil
		argsNum := len(args)

		if argsNum > 0 {
			fileHandler = args[0].(*FileHandler)
		} else {
			fileHandler, err = NewFileHandler("./logs", "error.log",
				"daily", 7)
			if err != nil {
				return nil, err
			}
		}

		output = fileHandler
	} else {
		output = os.Stdout
	}

	lg := log.New(output, "", log.LstdFlags|log.Lshortfile)

	return &Logger{appender: a, level: NewLevel(level),
		logger: lg, fileHandler: fileHandler, name: name}, nil
}

func (l *Logger) sync() {
	if l.fileHandler != nil && l.fileHandler.fd != nil {
		l.fileHandler.fd.Sync()
	}
}

func (l *Logger) write(level Level, message string) {
	if level < l.level {
		return
	}
	l.logger.SetPrefix(fmt.Sprintf("[%s][%s]", l.name, level.String()))
	l.logger.Output(2, message)
}

func (l *Logger) output(level Level, v ...interface{}) {
	l.write(level, fmt.Sprint(v...))
	if level == FATAL {
		l.sync()
		os.Exit(1)
	}
}

func (l *Logger) outputf(level Level, format string, v ...interface{}) {
	l.write(level, fmt.Sprintf(format, v...))
}

// Trace log
func (l *Logger) Trace(v ...interface{}) {
	l.output(TRACE, v)
}

// Tracef log
func (l *Logger) Tracef(format string, v ...interface{}) {
	l.outputf(TRACE, format, v)
}

// Debug log
func (l *Logger) Debug(v ...interface{}) {
	l.output(DEBUG, v)
}

// Debugf log
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.outputf(DEBUG, format, v)
}

// Info log
func (l *Logger) Info(v ...interface{}) {
	l.output(INFO, v)
}

// Infof log
func (l *Logger) Infof(format string, v ...interface{}) {
	l.outputf(INFO, format, v)
}

// Warning log
func (l *Logger) Warning(v ...interface{}) {
	l.output(WARNING, v)
}

// Warningf log
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.outputf(WARNING, format, v)
}

func (l *Logger) Error(v ...interface{}) {
	l.output(ERROR, v)
}

//Errorf log
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.outputf(ERROR, format, v)
}

// Fatal log
func (l *Logger) Fatal(v ...interface{}) {
	l.output(FATAL, v)
}

// Fatalf log
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.outputf(FATAL, format, v)
}
