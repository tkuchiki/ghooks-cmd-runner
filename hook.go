package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type hook struct {
	Event          string   `toml:"event"`
	Cmd            string   `toml:"command"`
	Branch         string   `toml:"branch"`
	IncludeActions []string `toml:"include_actions"`
	ExcludeActions []string `toml:"exclude_actions"`
	AccessToken    string   `toml:"access_token"`
	isEncoded      bool
}

func (h hook) setIsEncoded(b bool) {
	h.isEncoded = b
}

func (h hook) callback(payload interface{}) {
	action := parseAction(payload)

	if !includeActions(action, h.IncludeActions) {
		log.Info(fmt.Sprintf("skipped action = %s (%s)", action, h.Event))
		return
	}

	if excludeActions(action, h.ExcludeActions) {
		log.Info(fmt.Sprintf("skipped action = %s (%s)", action, h.Event))
		return
	}

	branch := parseBranch(payload)
	var matched bool
	matched, err := matchBranch(branch, h.Branch)
	if err != nil {
		log.Fatal(err)
	}

	if matched {
		m := new(sync.Mutex)

		var buf []byte
		buf, err = json.Marshal(payload)
		if err != nil {
			log.Fatal(err)
		}

		m.Lock()

		var g githubClient
		if h.Event == "pull_request" && h.isNotBlankAccessToken() {
			owner, repo, ref := parsePullRequestStatus(payload)
			g = NewClient(owner, repo, ref, h.AccessToken)

			g.pendingStatus()
		}

		var successF, failureF *os.File
		var successDir, failureDir string
		if h.Event == "pull_request" && h.isNotBlankAccessToken() {
			successF, successDir, err = createTempFile()
			if err != nil {
				log.Error(err)
				m.Unlock()
				removeDirs(successDir)
				return
			}
			os.Setenv("SUCCESS_TARGET_FILE", successF.Name())

			failureF, failureDir, err = createTempFile()
			if err != nil {
				log.Error(err)
				m.Unlock()
				removeDirs(successDir, failureDir)
				return
			}
			os.Setenv("FAILURE_TARGET_FILE", failureF.Name())
		}

		err = runCmd(h.Cmd, buf, h.isEncoded)

		if h.Event == "pull_request" && h.isNotBlankAccessToken() && err == nil {
			err = g.successStatus(readlineTempFile(successF))
			removeDirs(successDir, failureDir)
			if err != nil {
				log.Error(err)
				m.Unlock()
				removeDirs(successDir, failureDir)
				return
			}
		}

		if err != nil {
			if h.isNotBlankAccessToken() {
				err = g.failureStatus(readlineTempFile(failureF))
				removeDirs(successDir, failureDir)
				if err != nil {
					log.Error(err)
				}
			}
			log.Error(err)
			m.Unlock()
			return
		}

		time.Sleep(1000 * time.Millisecond)
		m.Unlock()
	}
}
