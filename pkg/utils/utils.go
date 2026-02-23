package utils


// FileName Operation
func GetExtOfFile(fileName string) string {
	for i := len(fileName)-1 ; i >= 0 ; i -- {
		if (fileName[i] == '.') {
			return fileName[i+1:]
		}
	}

	return ""
}
