package utils

import "os"

func MustFileName(fileName string) (created bool) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		created = true
		file, err := os.Create(fileName)
		if err == nil {
			err = file.Close()
		}
	}
	if err != nil {
		panic(err)
	}
	return
}
