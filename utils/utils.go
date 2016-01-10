package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/labstack/gommon/log"
	"io/ioutil"
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

func UpdateFile(file, val string) error {
	dir := filepath.Dir(file)
	log.Info("creating dir %s", dir)
	dir, err := CreateDirIfReqd(dir)
	if err != nil {
		return err
	}
	ioutil.WriteFile(filepath.Join(dir, filepath.Base(file)), []byte(val), 0777)
	return nil
}

func DefaultDir(path string) string {
	defaultDir := "."
	if GOPATH := os.Getenv("GOPATH"); GOPATH != "" {
		srcDir, err := filepath.Abs(filepath.Join(GOPATH, path))
		if err == nil {
			defaultDir = srcDir
		}
	}
	return defaultDir
}
