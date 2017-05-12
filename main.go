package main

import (
	"fmt"
	"os"

	"github.com/Konboi/ghooks"
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
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
	isEncoded   = kingpin.Flag("no-base64", "use raw payload").Default("false").Bool()
	log         = logrus.New()
)

func main() {
	kingpin.CommandLine.Help = "Receives Github webhooks and runs commands"
	kingpin.Version("0.3.2")
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
		h.setIsEncoded(!*isEncoded)
		hooks.On(h.Event, h.callback)
	}

	hooks.Run()
}
