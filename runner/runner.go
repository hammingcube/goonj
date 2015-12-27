package runner

import (
	"github.com/labstack/gommon/log"
	"github.com/maddyonline/hey/cmd/build"
	"io/ioutil"
	"path/filepath"
)

func RunIt(src, lang string) ([]byte, error) {
	log.Info("In RunIt, got src=%s, lang=%s", src, lang)
	outFile := "prog-output.txt"
	outFile = filepath.Join(filepath.Dir(src), outFile)

	_, stderr, err := build.RunFunc(&build.Options{
		Src:         src,
		OutFile:     outFile,
		DryRun:      false,
		Language:    lang,
		OnlyCompile: false,
	})
	/*
		_, stderr, err := build.RunFunc(src, outFile, dryRun)
	*/
	if err != nil {
		return []byte(stderr), err
	}

	progOutput, err := ioutil.ReadFile(outFile)
	return progOutput, err
}
