package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func runCmd(command, payload string) error {
	var out []byte
	var err error

	os.Setenv("GITHUB_WEBHOOK_PAYLOAD", payload)
	cmds := strings.Fields(command)
	if len(cmds) > 1 {
		out, err = exec.Command(cmds[0], cmds[1:]...).CombinedOutput()
	} else {
		out, err = exec.Command(cmds[0]).CombinedOutput()
	}

	if len(out) > 0 {
		outputLines(out)
	}

	if err != nil {
		log.Println(err)
	}

	return err
}

func outputLines(data []byte) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		log.Println(scanner.Text())
	}
}

func openFile(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func createPIDFile(filename string) error {
	return ioutil.WriteFile(filename, []byte(fmt.Sprint(os.Getpid())), 0644)
}

func matchBranch(branch, pattern string) (bool, error) {
	if branch == "" {
		return true, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	return re.Match([]byte(branch)), nil
}

func createTempFile() (*os.File, string, error) {
	var f *os.File
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return f, "", err
	}

	f, err = ioutil.TempFile(dir, "ghooks-cmd-runner")

	return f, dir, err
}

func readlineTempFile(f *os.File) string {
	scanner := bufio.NewScanner(f)
	_ = scanner.Scan()
	return scanner.Text()
}

func removeDirs(files ...string) {
	for _, f := range files {
		os.RemoveAll(f)
	}
}
