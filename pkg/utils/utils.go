package utils

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
)

var (
	unknown  	= []byte("???")
	centerDot	= []byte("")
	dot			= []byte(".")
	slash		= []byte("/")
)

// FileName Operation
func GetExtOfFile(fileName string) string {
	for i := len(fileName)-1 ; i >= 0 ; i -- {
		if (fileName[i] == '.') {
			return fileName[i+1:]
		}
	}

	return ""
}

func IsFileExists(dirPath string) bool {
	_, err := os.Stat(dirPath)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}

	return true
}


// Time Operation
const (
	TimeFormatSecond = "2006-01-02 15:04:05"
	TimeFormatMinute = "2006-01-02 15:04"
	TimeFormatDateV1 = "2006-01-02"
	TimeFormatDateV2 = "2006_01_02"
	TimeFormatDateV3 = "20060102150405"
	TimeFormatDateV4 = "2006/01/02-15:04:05.000"
)

func FormatTime(t time.Time) string {
	var tStr = t.Format(TimeFormatSecond)
	return tStr
}

// UUID
// @Description: No Need Distributed ID Generator
func GenerateUUID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

// IP
// @Description: Get Local IP
func GetLocalIP() (net.IP, error) {
	tables, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, table := range tables {
		addrs, err := table.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}

			if v4 := ipnet.IP.To4() ; v4 != nil {
				return v4, nil
			}
		}
	}

	return nil, fmt.Errorf("Cannot Find Local IP Address")
}


// Panic Handler
// @Description: Panic Trace
func PanicTrace(err any) string {
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false)

	return fmt.Sprintf("panic %v %s", err, string(stackBuf[:n]))
}

// @Description: Transform From Panic To Error
// 				 Swap Func And Recover 
func PanicToError(f func()) (err error) {
	defer func() {
		if e := recover() ; e != nil {
			err = fmt.Errorf("%s", PanicTrace(e))
		}
	}()

	f()
	return 
}

// Panic Handler Utils

// Error Handler
// @Description: Buffer Caller Stack
func Stack(skip int) []byte {
	buf := &bytes.Buffer{}
	var lines [][]byte
	var lastFile string
	for i := skip; ; i ++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)

		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}

	return buf.Bytes()
}

func source(lines [][]byte, n int) []byte {
	n--
	if n < 0 || n >= len(lines) {
		return unknown
	}
	return bytes.TrimSpace(lines[n])
}

func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return unknown
	}

	name := []byte(fn.Name())
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}

	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}

	name = bytes.ReplaceAll(name, centerDot, dot)
	return name
}

// Encrypt Tool
func MD5(s string) string {
	data := []byte(s)
	hash := md5.Sum(data)
	encryptedStr := fmt.Sprintf("%x", hash)
	return encryptedStr
}














