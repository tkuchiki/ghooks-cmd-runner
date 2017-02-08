package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

var (
	stdouterrCh = make(chan string)
	quitCh      = make(chan struct{})
)

func runCmd(command, payload string) error {
	var cmd *exec.Cmd

	os.Setenv("GITHUB_WEBHOOK_PAYLOAD", payload)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	stdin, stdinErr := cmd.StdinPipe()
	if stdinErr != nil {
		return stdinErr
	}

	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		return stdoutErr
	}

	stderr, stderrErr := cmd.StderrPipe()
	if stderrErr != nil {
		return stderrErr
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		_, werr := io.WriteString(stdin, payload)
		if perr, ok := werr.(*os.PathError); ok && perr.Err == syscall.EPIPE {
			// ignore EPIPE
		} else if werr != nil {
			log.Println(werr)
		}
		stdin.Close()
		wg.Done()
	}()

	go func() {
		readIo(stdout, stdouterrCh)
		stdout.Close()
		wg.Done()
	}()

	go func() {
		readIo(stderr, stdouterrCh)
		stderr.Close()
		wg.Done()
	}()

	go func() {
		for {
			select {
			case txt := <-stdouterrCh:
				log.Println(txt)
			case <-quitCh:
				break
			}
		}
	}()

	wg.Wait()

	err := cmd.Wait()
	quitCh <- struct{}{}

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

func readIo(r io.Reader, q chan string) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		q <- scanner.Text()
	}
}
