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

func parseBranch(payload interface{}) string {
	j := payload.(map[string]interface{})
	if _, ok := j["ref"]; !ok {
		return ""
	}

	branches := strings.SplitN(j["ref"].(string), "/", 3)

	if len(branches) == 3 {
		return branches[2]
	}

	return ""
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
