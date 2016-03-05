package main

import (
	"github.com/go-ini/ini"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
)

var config = struct {
	Port, Quality   int
	Source, Channel string
}{}

func main() {
	readConfig()
	go fallback()
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

// Dispatcher writes the received []byte from either the fallback stream or the
// transcoded livestreamer output to all connected clients
type Dispatcher struct {
	isLive bool
	sync.Mutex
	clients  map[int64]chan<- []byte
	fallback chan []byte
	stream   chan []byte
}

// Listen receives messages from
func (d *Dispatcher) Listen() {
	for {
		select {
		case buf := <-d.fallback:
			if !d.isLive {
				d.write(buf)
			}
		case buf := <-d.stream:
			d.write(buf)
		}
	}
}

// Add a client to receive messages from the dispatcher
func (d *Dispatcher) Add(cl *Client) {
	d.Lock()
	for {
		id := rand.Int63()
		if _, ok := d.clients[id]; !ok {
			d.clients[id] = cl.writer
			cl.id = id
			break
		}
	}
	d.Unlock()
}

// Remove removes a client from the set of listening clients
func (d *Dispatcher) Remove(cl *Client) {
	d.Lock()
	delete(d.clients, cl.id)
	d.Unlock()
}

func (d *Dispatcher) write(buf []byte) {
	d.Lock()
	for _, w := range d.clients {
		w <- buf
	}
	d.Unlock()
}

// Client is a client connected through HTTP, that is being sent the audio
// stream.
type Client struct {
	id     int64
	writer chan []byte
	ip     string
}

func fallback() {
	for {

	}
}
