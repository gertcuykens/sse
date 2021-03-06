package sse

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Conn SSE
type Conn struct {
	Stream chan Event
	Err    error
	Retry  int
	close  chan struct{}
}

// Event SSE
type Event struct {
	ID   string
	Type string
	Data []byte
}

// NewConn SSE
func NewConn(url string, id string) (*Conn, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Last-Event-ID", id)
	req.Close = true
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	xhr := &http.Client{Transport: tr}
	res, err := xhr.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Println("Client Connected.")

	conn := &Conn{
		Stream: make(chan Event, 1024),
		close:  make(chan struct{}),
	}

	go func() {
		<-conn.close
		res.Body.Close()
		close(conn.Stream)
		fmt.Println("Client Closed.")
	}()

	go stream(res.Body, conn)

	return conn, nil
}

// Close Conn.
func (c *Conn) Close() error {
	c.close <- struct{}{}
	return nil
}

func stream(body io.Reader, conn *Conn) {
	s := bufio.NewScanner(body)
	s.Split(bufio.ScanLines)

	var event Event
	for s.Scan() {
		line := s.Bytes()
		if len(line) == 0 {
			conn.Stream <- event
			event = Event{}
		}
		field := line
		value := []byte{}
		if colon := bytes.IndexByte(line, ':'); colon != -1 {
			if colon == 0 {
				continue
			}
			field = line[:colon]
			value = line[colon+1:]
			if value[0] == ' ' {
				value = value[1:]
			}
		}
		switch string(field) {
		case "event":
			event.Type = string(value)
		case "data":
			event.Data = append(append(event.Data, value...), '\n')
		case "id":
			event.ID = string(value)
		case "retry":
			if i, err := strconv.Atoi(string(value)); err != nil {
				conn.Retry = i
			}
		default:
			// Ignored
		}
	}

	conn.Err = s.Err()
	conn.Close()
}
