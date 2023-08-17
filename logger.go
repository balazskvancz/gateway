package gateway

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

type (
	fileType     string
	logTypeName  string
	logTypeValue uint8
)

const (
	fInfo  fileType = "info"
	fError fileType = "error"

	tInfo    logTypeName = "info"
	tWarning logTypeName = "warning"
	tError   logTypeName = "error"
)

const (
	vInfo logTypeValue = 1 << iota
	vWarning
	vError
)

const defaultLogLevel = vInfo + vWarning + vError

var logLevelValues = map[logTypeName]logTypeValue{
	tInfo:    vInfo,
	tWarning: vWarning,
	tError:   vError,
}

type fileLogger struct {
	*log.Logger
	f          *os.File
	latestDate string
	fType      fileType
}

type gatewayLogger struct {
	logLevel logTypeValue
	*log.Logger
	fileLoggers map[fileType]*fileLogger
}

type logger interface {
	Info(string)
	Error(string)
	Warning(string)
	clean()
	disable(logTypeValue)
}

const (
	defaultLogFlag int = log.LstdFlags
)

var (
	logPrefix = fmt.Sprintf("[api-gateway %s] ", Version)
)

var _ logger = (*gatewayLogger)(nil)

func newGatewayLogger() logger {
	fileLoggers := make(map[fileType]*fileLogger)

	fileLoggers[fInfo] = newFileLogger(fInfo, logPrefix, defaultLogFlag)

	return &gatewayLogger{
		Logger:      log.New(os.Stdout, logPrefix, defaultLogFlag),
		fileLoggers: fileLoggers,
		logLevel:    defaultLogLevel,
	}
}

func newFileLogger(t fileType, prefix string, flag int) *fileLogger {
	f := getFile(t)
	if f == nil {
		return nil
	}

	l := &fileLogger{
		f:          f,
		latestDate: getCurrentDate(),
		fType:      t,
	}

	logger := log.New(l.f, prefix, flag)
	l.Logger = logger

	return l
}

func (l *gatewayLogger) Info(v string) {
	l.write(tInfo, fmt.Sprintf("[INFO] – %s\n", v), fInfo)
}

func (l *gatewayLogger) Error(v string) {
	l.write(tError, fmt.Sprintf("[ERROR] – %s\n", v), fError)
}

func (l *gatewayLogger) Warning(v string) {
	l.write(tWarning, fmt.Sprintf("[WARNING] – %s\n", v), fInfo)
}

func (l *gatewayLogger) write(logType logTypeName, t string, fType fileType) {
	if logLevelValues[logType]&l.logLevel == 0 {
		return
	}
	l.Printf(t)
	fileLogger, exits := l.fileLoggers[fType]
	if !exits {
		newFileLogger := newFileLogger(fType, logPrefix, defaultLogFlag)
		l.fileLoggers[fType] = newFileLogger
		fileLogger = newFileLogger
	}
	fileLogger.write(t)
}

func (l *gatewayLogger) clean() {
	for _, v := range l.fileLoggers {
		v.clean()
	}
}

func (fl *fileLogger) write(t string) {
	if fl == nil {
		return
	}

	d := getCurrentDate()

	if d != fl.latestDate {
		f := getFile(fl.fType)
		if f == nil {
			return
		}

		fl.f.Close()

		fl.f = f
		fl.latestDate = d
		fl.Logger = log.New(fl.f, logPrefix, defaultLogFlag)
	}

	fl.Printf(t)
}

func (fl *fileLogger) clean() {
	if fl == nil {
		return
	}
	fl.f.Close()
}

func getFile(ty fileType) *os.File {
	var (
		fname = getFileName(ty)
		fpath = path.Join("logs", fname)
	)

	f, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		return f
	}

	if !errors.Is(err, os.ErrNotExist) {
		fmt.Println(err)
		return nil
	}

	f, err = os.Create(fpath)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return f
}

func getFileName(ft fileType) string {
	dateStr := getCurrentDate()

	return fmt.Sprintf("api-gateway-%s-%s.log", dateStr, ft)
}

func getCurrentDate() string {
	return time.Now().Format("2006_01_02")
}

func (l *gatewayLogger) disable(d logTypeValue) {
	if l.logLevel&d == 0 {
		return
	}
	l.logLevel -= d
}
