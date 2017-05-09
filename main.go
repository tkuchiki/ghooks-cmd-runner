package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Konboi/ghooks"
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"sync"
	"time"
)

type cmd struct {
	command string
	payload string
}

var (
	defaultPort = 18889
	defaultHost = "127.0.0.1"
	file        = kingpin.Flag("config", "config file location").Short('c').Required().String()
	port        = kingpin.Flag("port", "listen port").Short('p').Default(fmt.Sprint(defaultPort)).Int()
	host        = kingpin.Flag("host", "listen host").Default(defaultHost).String()
	logfile     = kingpin.Flag("logfile", "log file location").Short('l').String()
	pidfile     = kingpin.Flag("pidfile", "pid file location").String()
	log         = logrus.New()
)

func main() {
	kingpin.CommandLine.Help = "Receives Github webhooks and runs commands"
	kingpin.Version("0.3.1")
	kingpin.Parse()

	tmpConf := config{
		Port:    *port,
		Host:    *host,
		Logfile: *logfile,
		Pidfile: *pidfile,
	}

	conf, err := loadToml(*file, tmpConf)
	if err != nil {
		log.Fatal(err)
	}

	if conf.Logfile != "" {
		var f *os.File
		f, err = openFile(conf.Logfile)
		defer f.Close()

		if err != nil {
			log.Fatal(err)
		}
		log.Out = f
		log.Formatter = &logrus.TextFormatter{DisableColors: true}
	}

	hooks := ghooks.NewServer(conf.Port, conf.Host)

	if conf.Secret != "" {
		hooks.Secret = conf.Secret
	}

	if conf.Pidfile != "" {
		err = createPIDFile(conf.Pidfile)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, h := range conf.Hook {
		if h.Event == "" {
			log.Fatal("event is required.")
		}
		hookingBranch := h.Branch
		hooks.On(h.Event, func(payload interface{}) {
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
			matched, err = matchBranch(branch, hookingBranch)
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

				p := base64.StdEncoding.EncodeToString(buf)
				err = runCmd(h.Cmd, p)

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
		})
	}

	hooks.Run()
}
