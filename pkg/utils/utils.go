package utils

import "os"

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

