package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

var (
	stdouterrCh = make(chan string)
	quitCh      = make(chan struct{})
)

func runCmd(command string, buf []byte) error {
	var cmd *exec.Cmd

	payload := base64.StdEncoding.EncodeToString(buf)
	b := bytes.NewBuffer(buf)
	os.Setenv("GITHUB_WEBHOOK_PAYLOAD", payload)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	go readIo(stdout)
	go readIo(stderr)

	go func() {
		defer stdin.Close()
		_, werr := b.WriteTo(stdin)
		if perr, ok := werr.(*os.PathError); ok && perr.Err == syscall.EPIPE {
			// ignore EPIPE
		} else if werr != nil {
			log.Println(werr)
		}
	}()

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
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

func readIo(r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		log.Info(scanner.Text())
	}
}
