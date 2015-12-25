package runner

import (
	"github.com/maddyonline/hey/cmd/build"
	"io/ioutil"
	"path/filepath"
)

func RunIt(src string) ([]byte, error) {
	//src := "/Users/madhavjha/src/github.com/maddyonline/workdir/main.cpp"
	outFile := "prog-output.txt"
	outFile = filepath.Join(filepath.Dir(src), outFile)
	dryRun := false
	_, stderr, err := build.RunFunc(src, outFile, dryRun)
	if err != nil {
		return []byte(stderr), err
	}
	progOutput, err := ioutil.ReadFile(outFile)
	return progOutput, err
}
