package main

import (
	"github.com/go-ini/ini"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

var config = struct {
	Quality                int
	Address, URL, Fallback string
}{}

var dispatcher = Dispatcher{
	clients:  make(map[int64]chan<- []byte),
	fallback: make(chan []byte),
	stream:   make(chan []byte),
}

func main() {
	readConfig()
	go fallback()
	go stream()
	go dispatcher.Listen()
	log.Printf("Listening on %s", config.Address)
	throw(http.ListenAndServe(config.Address, http.HandlerFunc(serveStream)))
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

var headers = map[string]string{
	"Access-Control-Allow-Origin": "*",
	"content-type":                "audio/aac",
	"Expires":                     "Thu, 01 Jan 1970 00:00:00 GMT",
	"Cache-Control":               "no-cache, no-store",
	"Transfer-Encoding":           "chunked",
	"Connection":                  "keep-alive",
}

func serveStream(res http.ResponseWriter, req *http.Request) {
	cl := &Client{
		ip:     req.RemoteAddr,
		writer: make(chan []byte),
	}
	dispatcher.Add(cl)

	head := res.Header()
	for key, val := range headers {
		head.Set(key, val)
	}

	for {
		_, err := res.Write(<-cl.writer)
		if err != nil {
			dispatcher.Remove(cl)
			break
		}
	}
}

// Dispatcher writes the received []byte from either the fallback stream or the
// transcoded livestreamer output to all connected clients
type Dispatcher struct {
	isLive      bool
	lastMessage time.Time
	sync.Mutex
	clients  map[int64]chan<- []byte
	fallback chan []byte
	stream   chan []byte
}

// Listen receives messages from both the stream and fallback loop
func (d *Dispatcher) Listen() {
	for {
		select {
		case buf := <-d.fallback:
			// If more than  100 ms have passed since the last stream message,
			// switch to the fallback
			if d.isLive {
				if time.Since(d.lastMessage).Nanoseconds() > 100000 {
					d.isLive = false
					d.write(buf)
				}
			} else {
				d.write(buf)
			}
		case buf := <-d.stream:
			d.isLive = true
			d.lastMessage = time.Now()
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
	log.Printf("%s connected\n", cl.ip)
	d.printClientCount()
	d.Unlock()
}

func (d *Dispatcher) printClientCount() {
	log.Printf("%v total clients connected\n", len(d.clients))
}

// Remove removes a client from the set of listening clients
func (d *Dispatcher) Remove(cl *Client) {
	d.Lock()
	delete(d.clients, cl.id)
	log.Printf("%s disconnected\n", cl.ip)
	d.printClientCount()
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

// Play the fallback MP4 in a loop and streams it to the dispatcher
func fallback() {
	for {
		ffmpeg := exec.Command(
			"ffmpeg", "-v", "error",
			"-re", "-i", config.Fallback,
			"-c:a", "copy", "-f", "adts", "-bufsize", "496K", "-",
		)
		ffmpeg.Stderr = os.Stderr
		ffmpeg.Stdout = streamWriter{true}
		throw(ffmpeg.Run())
	}
}

// Streamwriter pipes the output of a child process into the dispatcher
type streamWriter struct {
	isFallback bool
}

func (s streamWriter) Write(buf []byte) (int, error) {
	if s.isFallback {
		dispatcher.fallback <- buf
	} else {
		dispatcher.stream <- buf
	}
	return len(buf), nil
}

func stream() {
	for {
		ls := exec.Command(
			"livestreamer",
			"--retry-streams", "10",
			"-O",
			config.URL, "best",
		)
		ls.Stderr = os.Stderr
		ls.Stdout = streamWriter{}
		throw(ls.Run())
	}
}
