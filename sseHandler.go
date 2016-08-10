package sse

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Handler SSE.
type Handler struct {
	Push        chan []byte
	Retry       chan<- int
	newClient   chan chan<- []byte
	closeClient chan chan<- []byte
	clients     map[chan<- []byte]int
}

// NewHandler SSE.
func NewHandler() (handler *Handler) {
	handler = &Handler{
		Push:        make(chan []byte),
		Retry:       make(chan<- int),
		newClient:   make(chan chan<- []byte),
		closeClient: make(chan chan<- []byte),
		clients:     make(map[chan<- []byte]int),
	}
	go handler.listen()
	return
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "No stream!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan []byte)
	handler.newClient <- messageChan

	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		handler.closeClient <- messageChan
	}()

	for {
		//fmt.Fprint(w, string(<-messageChan))
		io.Copy(w, bytes.NewReader(<-messageChan))
		flusher.Flush()
	}

}

func (handler *Handler) listen() {
	for {
		select {
		case c := <-handler.newClient:
			handler.clients[c] = len(handler.clients)
			fmt.Printf("IP<-: %d\n", len(handler.clients))
		case c := <-handler.closeClient:
			delete(handler.clients, c)
			fmt.Printf("IP->: %d\n", len(handler.clients))
		case e := <-handler.Push:
			for clientEventChan := range handler.clients {
				clientEventChan <- e
			}
		}
	}
}
