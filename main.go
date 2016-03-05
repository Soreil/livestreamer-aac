package main

import (
	"github.com/go-ini/ini"
	"io/ioutil"
	"os"
)

var config = struct {
	Port, Quality   int
	Source, Channel string
}{}

func main() {
	readConfig()
}

func throw(err error) {
	if err != nil {
		panic(err)
	}
}

func readConfig() {
	const (
		path = "./config.ini"
		def  = "./defaults.ini"
	)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			buf, err = ioutil.ReadFile(def)
			throw(err)
			throw(ioutil.WriteFile(path, buf, 0600))
		} else {
			panic(err)
		}
	}
	c, err := ini.Load(buf)
	throw(err)
	throw(c.MapTo(&config))
}
