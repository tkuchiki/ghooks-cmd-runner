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
	kingpin.Version("0.2.0")
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

		hooks.On(h.Event, func(payload interface{}) {
			branch := parseBranch(payload)
			var matched bool
			matched, err = matchBranch(branch, h.Branch)
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

				encPayload := base64.StdEncoding.EncodeToString(buf)

				m.Lock()
				err = runCmd(h.Cmd, encPayload)
				if err != nil {
					log.Fatal(err)
				}
				m.Unlock()
			}
		})
	}

	hooks.Run()
}
