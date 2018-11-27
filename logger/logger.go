package logger

//	"log"

const (
	//go-logger version
	_VER string = "1.0.3"
)

type LEVEL int32
type UNIT int64
type ROLLTYPE int //dailyRolling ,rollingFile

const _DATEFORMAT = "2006-01-02"

var logLevel LEVEL = 1

const (
	_       = iota
	KB UNIT = 1 << (iota * 10)
	MB
	GB
	TB
)

const (
	ALL LEVEL = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

const (
	DAILY ROLLTYPE = iota
	ROLLFILE
)

func SetConsole(isConsole bool) {
	defaultlog.setConsole(isConsole)
}

func SetLevel(_level LEVEL) {
	defaultlog.setLevel(_level)
}

func SetFormat(logFormat string) {
	defaultlog.setFormat(logFormat)
}

func SetRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT) {
	defaultlog.setRollingFile(fileDir, fileName, maxNumber, maxSize, _unit)
}

func SetRollingDaily(fileDir, fileName string) {
	defaultlog.setRollingDaily(fileDir, fileName)
}

func Debug(v ...interface{}) {
	defaultlog.debug(v...)
}
func Info(v ...interface{}) {
	defaultlog.info(v...)
}
func Warn(v ...interface{}) {
	defaultlog.warn(v...)
}
func Error(v ...interface{}) {
	defaultlog.error(v...)
}
func Fatal(v ...interface{}) {
	defaultlog.fatal(v...)
}

func SetLevelFile(level LEVEL, dir, fileName string) {
	defaultlog.setLevelFile(level, dir, fileName)
}
