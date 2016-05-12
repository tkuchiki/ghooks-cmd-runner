package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Konboi/ghooks"
	"github.com/Sirupsen/logrus"
	"github.com/VividCortex/godaemon"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"os/exec"
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
		log.Println(string(out))
	}

	if err != nil {
		log.Println(err)
	}

	return err
}

func openFile(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func createPIDFile(filename string) error {
	return ioutil.WriteFile(filename, []byte(fmt.Sprint(os.Getpid())), 0644)
}

var (
	defaultPort = 18889
	defaultHost = "127.0.0.1"
	file        = kingpin.Flag("config", "config file location").Short('c').Required().String()
	port        = kingpin.Flag("port", "listen port").Short('p').Default(fmt.Sprint(defaultPort)).Int()
	host        = kingpin.Flag("host", "listen host").Default(defaultHost).String()
	logfile     = kingpin.Flag("logfile", "log file location").Short('l').String()
	pidfile     = kingpin.Flag("pidfile", "pid file location").String()
	daemon      = kingpin.Flag("daemon", "enable daemon mode").Short('d').Bool()
	log         = logrus.New()
)

func main() {
	kingpin.CommandLine.Help = "Receives Github webhooks and runs commands"
	kingpin.Version("0.1.0")
	kingpin.Parse()

	tmpConf := config{
		Port:      *port,
		Host:      *host,
		Logfile:   *logfile,
		Daemonize: *daemon,
		Pidfile:   *pidfile,
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

	if conf.Daemonize {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	}

	for _, h := range conf.Hook {
		hooks.On(h.Event, func(payload interface{}) {
			var buf []byte
			buf, err = json.Marshal(payload)
			if err != nil {
				log.Fatal(err)
			}

			encPayload := base64.StdEncoding.EncodeToString(buf)
			err = runCmd(h.Cmd, encPayload)
			if err != nil {
				log.Fatal(err)
			}
		})
	}

	hooks.Run()
}
