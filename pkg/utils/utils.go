package utils

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/google/uuid"
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
			ipnet, ok := addr.(*net.IPAddr)
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

	return fmt.Sprintf("panic %v %s", err, stackBuf[:n])
}

// @Description: Transform From Panic To Error
// 				 Swap Func And Recover 
func PanicToError(f func()) (err error) {
	defer func() {
		if e := recover() ; e != nil {
			err = fmt.Errorf(PanicTrace(e))
		}
	}()

	f()
	return 
}




