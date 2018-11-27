package logger

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var defaultlog *logBean = getdefaultLogger()
var skip int = 4

type logger struct {
	lb *logBean
}

func (this *logger) SetConsole(isConsole bool) {
	this.lb.setConsole(isConsole)
}

func (this *logger) SetLevel(_level LEVEL) {
	this.lb.setLevel(_level)
}

func (this *logger) SetFormat(logFormat string) {
	this.lb.setFormat(logFormat)
}

func (this *logger) SetRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT) {
	this.lb.setRollingFile(fileDir, fileName, maxNumber, maxSize, _unit)
}

func (this *logger) SetRollingDaily(fileDir, fileName string) {
	this.lb.setRollingDaily(fileDir, fileName)
}

func (this *logger) Debug(v ...interface{}) {
	this.lb.debug(v...)
}
func (this *logger) Info(v ...interface{}) {
	this.lb.info(v...)
}
func (this *logger) Warn(v ...interface{}) {
	this.lb.warn(v...)
}
func (this *logger) Error(v ...interface{}) {
	this.lb.error(v...)
}
func (this *logger) Fatal(v ...interface{}) {
	this.lb.fatal(v...)
}

func (this *logger) SetLevelFile(level LEVEL, dir, fileName string) {
	this.lb.setLevelFile(level, dir, fileName)
}

type logBean struct {
	mu              *sync.Mutex
	logLevel        LEVEL
	maxFileSize     int64
	maxFileCount    int32
	consoleAppender bool
	rolltype        ROLLTYPE
	format          string
	id              string
	d, i, w, e, f   string //id
}

type fileBeanFactory struct {
	fbs map[string]*fileBean
	mu  *sync.RWMutex
}

var fbf = &fileBeanFactory{fbs: make(map[string]*fileBean, 0), mu: new(sync.RWMutex)}

func (this *fileBeanFactory) add(dir, filename string, _suffix int, maxsize int64, maxfileCount int32) {
	this.mu.Lock()
	defer this.mu.Unlock()
	id := md5str(fmt.Sprint(dir, filename))
	if _, ok := this.fbs[id]; !ok {
		this.fbs[id] = newFileBean(dir, filename, _suffix, maxsize, maxfileCount)
	}
}

func (this *fileBeanFactory) get(id string) *fileBean {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.fbs[id]
}

type fileBean struct {
	id           string
	dir          string
	filename     string
	_suffix      int
	_date        *time.Time
	mu           *sync.RWMutex
	logfile      *os.File
	lg           *log.Logger
	filesize     int64
	maxFileSize  int64
	maxFileCount int32
}

func GetLogger() (l *logger) {
	l = new(logger)
	l.lb = getdefaultLogger()
	return
}

func getdefaultLogger() (lb *logBean) {
	lb = &logBean{}
	lb.mu = new(sync.Mutex)
	lb.setConsole(true)
	return
}

func (this *logBean) setConsole(isConsole bool) {
	this.consoleAppender = isConsole
}

func (this *logBean) setLevelFile(level LEVEL, dir, fileName string) {
	key := md5str(fmt.Sprint(dir, fileName))
	switch level {
	case DEBUG:
		this.d = key
	case INFO:
		this.i = key
	case WARN:
		this.w = key
	case ERROR:
		this.e = key
	case FATAL:
		this.f = key
	default:
		return
	}
	var _suffix = 0
	if this.maxFileCount < 1<<31-1 {
		for i := 1; i < int(this.maxFileCount); i++ {
			if isExist(dir + "/" + fileName + "." + strconv.Itoa(i)) {
				_suffix = i
			} else {
				break
			}
		}
	}
	fbf.add(dir, fileName, _suffix, this.maxFileSize, this.maxFileCount)
}

func (this *logBean) setLevel(_level LEVEL) {
	this.logLevel = _level
}

func (this *logBean) setFormat(logFormat string) {
	this.format = logFormat
}

func (this *logBean) setRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if maxNumber > 0 {
		this.maxFileCount = maxNumber
	} else {
		this.maxFileCount = 1<<31 - 1
	}
	this.maxFileSize = maxSize * int64(_unit)
	this.rolltype = ROLLFILE
	mkdirlog(fileDir)
	var _suffix = 0
	for i := 1; i < int(maxNumber); i++ {
		if isExist(fileDir + "/" + fileName + "." + strconv.Itoa(i)) {
			_suffix = i
		} else {
			break
		}
	}
	this.id = md5str(fmt.Sprint(fileDir, fileName))
	fbf.add(fileDir, fileName, _suffix, this.maxFileSize, this.maxFileCount)
}

func (this *logBean) setRollingDaily(fileDir, fileName string) {
	this.rolltype = DAILY
	mkdirlog(fileDir)
	this.id = md5str(fmt.Sprint(fileDir, fileName))
	fbf.add(fileDir, fileName, 0, 0, 0)
}

func (this *logBean) console(v ...interface{}) {
	s := fmt.Sprint(v...)
	if this.consoleAppender {
		_, file, line, _ := runtime.Caller(skip)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		if this.format == "" {
			log.Println(file, strconv.Itoa(line), s)
		} else {
			vs := make([]interface{}, 0)
			vs = append(vs, file)
			vs = append(vs, strconv.Itoa(line))
			for _, vv := range v {
				vs = append(vs, vv)
			}
			log.Printf(fmt.Sprint("%s %s ", this.format, "\n"), vs...)
		}
	}
}

func (this *logBean) log(level string, v ...interface{}) {
	defer catchError()
	s := fmt.Sprint(v...)
	length := len([]byte(s))
	var lg *fileBean = fbf.get(this.id)
	var _level = ALL
	switch level {
	case "debug":
		if this.d != "" {
			lg = fbf.get(this.d)
		}
		_level = DEBUG
	case "info":
		if this.i != "" {
			lg = fbf.get(this.i)
		}
		_level = INFO
	case "warn":
		if this.w != "" {
			lg = fbf.get(this.w)
		}
		_level = WARN
	case "error":
		if this.e != "" {
			lg = fbf.get(this.e)
		}
		_level = ERROR
	case "fatal":
		if this.f != "" {
			lg = fbf.get(this.f)
		}
		_level = FATAL
	}
	if lg != nil {
		this.fileCheck(lg)
		lg.addsize(int64(length))
		if this.logLevel <= _level {
			if lg != nil {
				if this.format == "" {
					lg.write(level, s)
				} else {
					lg.writef(this.format, v...)
				}
			}
			this.console(v...)
		}
	} else {
		this.console(v...)
	}
}

func (this *logBean) debug(v ...interface{}) {
	this.log("debug", v...)
}
func (this *logBean) info(v ...interface{}) {
	this.log("info", v...)
}
func (this *logBean) warn(v ...interface{}) {
	this.log("warn", v...)
}
func (this *logBean) error(v ...interface{}) {
	this.log("error", v...)
}
func (this *logBean) fatal(v ...interface{}) {
	this.log("fatal", v...)
}

func (this *logBean) fileCheck(fb *fileBean) {
	defer catchError()
	if this.isMustRename(fb) {
		this.mu.Lock()
		defer this.mu.Unlock()
		if this.isMustRename(fb) {
			fb.rename(this.rolltype)
		}
	}
}

//--------------------------------------------------------------------------------

func (this *logBean) isMustRename(fb *fileBean) bool {
	switch this.rolltype {
	case DAILY:
		t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
		if t.After(*fb._date) {
			return true
		}
	case ROLLFILE:
		return fb.isOverSize()
	}
	return false
}

func (this *fileBean) nextSuffix() int {
	return int(this._suffix%int(this.maxFileCount) + 1)
}

func newFileBean(fileDir, fileName string, _suffix int, maxSize int64, maxfileCount int32) (fb *fileBean) {
	t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
	fb = &fileBean{dir: fileDir, filename: fileName, _date: &t, mu: new(sync.RWMutex)}
	fb.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	fb.lg = log.New(fb.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	fb._suffix = _suffix
	fb.maxFileSize = maxSize
	fb.maxFileCount = maxfileCount
	fb.filesize = fileSize(fileDir + "/" + fileName)
	fb._date = &t
	return
}

func (this *fileBean) rename(rolltype ROLLTYPE) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.close()
	nextfilename := ""
	switch rolltype {
	case DAILY:
		nextfilename = fmt.Sprint(this.dir, "/", this.filename, ".", this._date.Format(_DATEFORMAT))
	case ROLLFILE:
		nextfilename = fmt.Sprint(this.dir, "/", this.filename, ".", this.nextSuffix())
		this._suffix = this.nextSuffix()
	}
	if isExist(nextfilename) {
		os.Remove(nextfilename)
	}
	os.Rename(this.dir+"/"+this.filename, nextfilename)
	t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
	this._date = &t
	this.logfile, _ = os.OpenFile(this.dir+"/"+this.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	this.lg = log.New(this.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	this.filesize = fileSize(this.dir + "/" + this.filename)
}

func (this *fileBean) addsize(size int64) {
	atomic.AddInt64(&this.filesize, size)
}

func (this *fileBean) write(level string, v ...interface{}) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	s := fmt.Sprint(v...)
	this.lg.Output(skip+1, fmt.Sprintln(level, s))
}

func (this *fileBean) writef(format string, v ...interface{}) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	this.lg.Output(skip+1, fmt.Sprintf(format, v...))
}

func (this *fileBean) isOverSize() bool {
	return this.filesize >= this.maxFileSize
}

func (this *fileBean) close() {
	this.logfile.Close()
}

//-----------------------------------------------------------------------------------------------

func mkdirlog(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0666); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func fileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func md5str(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}

func catchError() {
	if err := recover(); err != nil {
		fmt.Println(string(debug.Stack()))
	}
}
