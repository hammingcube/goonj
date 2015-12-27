package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func RandId() string {
	size := 32 // change the length of the generated random string here

	rb := make([]byte, size)
	_, err := rand.Read(rb)

	if err != nil {
		fmt.Println(err)
	}

	rs := base64.URLEncoding.EncodeToString(rb)

	return rs
}

func CreateDirIfReqd(dir string) (string, error) {
	dirAbsPath, err := filepath.Abs(dir)
	if err != nil {
		return dirAbsPath, err
	}
	if _, err := os.Stat(dirAbsPath); err == nil {
		return dirAbsPath, nil
	}
	err = os.MkdirAll(dirAbsPath, 0777)
	return dirAbsPath, err
}
