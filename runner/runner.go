package runner

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/maddyonline/hey/cmd/build"
	"github.com/maddyonline/hey/cmd/judge"
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

type Options struct {
	DryRun   bool
	Raw      bool
	NoDocker bool
	Language string
}

func JudgeIt(src, lang string) ([]byte, error) {
	log.Info("In JudgeIt, got src=%s, lang=%s", src, lang)
	outFile := "judge-output.json"
	outFile = filepath.Join(filepath.Dir(src), outFile)
	judgeOutput, err := judge.RunFunc([]byte(PROB_CONFIG), &judge.Options{
		DryRun:   false,
		Language: lang,
	}, src, outFile)
	if err != nil {
		return []byte(fmt.Sprintf("Error: %v", err)), err
	}
	return []byte(judgeOutput.Status), nil
}

const PROB_CONFIG = `{
  "id": "abc12111",
  "original-source": "EPI",
  "primary-solution": {
    "url": "github.com/hammingcube/solutions",
    "path": "epi/prob-1/solutions/primary-solution",
    "src": "can_string_be_palindrome_soln2.cc",
    "lang": "cpp",
    "version": ""
  },
  "primary-generator": {
    "url": "github.com/hammingcube/solutions",
    "path": "epi/prob-1/generators/primary-generator",
    "src": "Can_string_be_palindrome_gen.cc",
    "lang": "cpp",
    "version": ""
  },
  "primary-runner": {
    "url": "github.com/hammingcube/solutions",
    "path": "runners/primary-runner",
    "src": "runtest.go",
    "lang": "go",
    "version": ""
  }
}`
