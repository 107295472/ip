
package util

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sync"
    "time"
    "os/exec"
    "strings"
)

type logger log.Logger

const saveLogTime int64 = 60 * 60 * 24 * 7

var mu sync.RWMutex
var fd *os.File
var logHandle *log.Logger
var logPath=GetCurrentPath()+"/log/"
func init() {
    if err := newLog(); err != nil {
        fmt.Printf("create log error %s\n", err)
    }
}
//截取字符串 start 起点下标 length 需要截取的长度
func Substr(str string, start int, length int) string {
    rs := []rune(str)
    rl := len(rs)
    end := 0

    if start < 0 {
        start = rl - 1 + start
    }
    end = start + length

    if start > end {
        start, end = end, start
    }

    if start < 0 {
        start = 0
    }
    if start > rl {
        start = rl
    }
    if end < 0 {
        end = 0
    }
    if end > rl {
        end = rl
    }

    return string(rs[start:end])
}

func GetCurrentPath() string {
    s, err := exec.LookPath(os.Args[0])
    if err != nil {
        fmt.Println(err.Error())
    }
    s = strings.Replace(s, "\\", "/", -1)
    s = strings.Replace(s, "\\\\", "/", -1)
    i := strings.LastIndex(s, "/")
    path := string(s[0 : i+1])
    return path
}
//判断文件是否存在
func IsNotExist(path string) bool {
    _, err := os.Stat(path)
    if err != nil {
        return true
    }

    return false
}
//生成文件夹
func Mkdir(path string) error {
    if IsNotExist(path) {
        err := os.MkdirAll(path, os.ModePerm)
        if err != nil {
            return err
        }
    }

    return nil
}
//获取 年月日
func GetYmd() string {
    return time.Now().Format("20060102")
}
//获取程序名称
func GetProgramName() string {
    return filepath.Base(os.Args[0])
}
func newLog() error {
    mu.Lock()
    defer mu.Unlock()

    if err := Mkdir(logPath); err != nil {
        fmt.Printf("mkdir error %s\n", err)
        return err
    }

    if fd != nil {
        fd.Close()
    }

    fileName := fmt.Sprintf("%s_%s.log",logPath+GetProgramName(), GetYmd())
    fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        fmt.Printf("create file error %s\n", err)
        return err
    }

    logHandle = log.New(fd, "", log.Ltime|log.Lshortfile)

    return nil
}

func LogInfo() *log.Logger {
    logHandle.SetPrefix("[info] ")
    return logHandle
}

func LogError() *log.Logger {
    logHandle.SetPrefix("[error] ")
    return logHandle
}

func LogWarning() *log.Logger {
    logHandle.SetPrefix("[warning] ")
    return logHandle
}

func LogDebug() *log.Logger {
    logHandle.SetPrefix("[debug] ")
    return logHandle
}

func (l *logger) Println(v ...interface{}) {
    mu.RLock()
    defer mu.RUnlock()

    logHandle.Println(v...)
}

func (l *logger) Printf(format string, v ...interface{}) {
    mu.RLock()
    defer mu.RUnlock()

    logHandle.Printf(format, v...)
}

func Errorf(format string, v ...interface{}) error {
    return fmt.Errorf(format, v...)
}

func Println(v ...interface{}) {
    fmt.Println(v...)
}

func Printf(format string, v ...interface{}) {
    fmt.Printf(format, v...)
}

// 给定时任务调用的函数，管理日常记录的日志
func CheckLog() {
    fileName := fmt.Sprintf("%s_%s.log", logPath+GetProgramName(), GetYmd())
    if IsNotExist(fileName) {
        if err := newLog(); err != nil {
            fmt.Printf("create log error %s\n", err)
        }
    }

    if err := filepath.Walk(logPath, walkFunc); err != nil {
        LogError().Printf("filePath %s walk error: %s\n", logPath, err)
    }
}

func walkFunc(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    }

    if info.IsDir() {
        return nil
    }

    if GetUnixTime()-info.ModTime().Unix() > saveLogTime {
        err := Rm(path)
        if err != nil {
            LogError().Printf("remove file [%s] error: %s\n", info.Name(), err)
            return err
        }
    }

    return nil
}
//输出时间戳
func GetUnixTime() int64 {
    return time.Now().Unix()
}
//删除指定目录或文件
func Rm(path string) error {
    err := os.Remove(path)
    if err != nil {
        return err
    }

    return nil
}