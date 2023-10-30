package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aimjel/minecraft"
	"github.com/aimjel/minecraft/packet"
	"net/http"
	"reflect"
	"strings"
)

type httpServer struct {
	srv *http.Server

	activeConns []*minecraft.Conn

	packets chan packet.Packet

	buf *bytes.Buffer

	enc *json.Encoder

	sb *strings.Builder
}

func startHttpServer() *httpServer {
	http.Handle("/", http.FileServer(http.Dir("web")))

	srv := &httpServer{
		srv:     &http.Server{Addr: "localhost:5000", Handler: http.DefaultServeMux},
		packets: make(chan packet.Packet, 128),
		buf:     bytes.NewBuffer(nil),
		sb:      &strings.Builder{},
	}
	srv.enc = json.NewEncoder(srv.buf)

	http.HandleFunc("/packets", srv.handlePackets)

	return srv
}

func (s *httpServer) handlePackets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	for pk := range s.packets {
		event, err := s.formatServerEvent("packet-list-update", pk)
		if err != nil {
			fmt.Println(err)
			break
		}

		_, err = fmt.Fprint(w, event)
		if err != nil {
			fmt.Println(err)
			break
		}

		flusher.Flush()
	}
}

func (s *httpServer) handleReceive(conn *minecraft.ProxyConn, pk packet.Packet, fromServer bool) bool {
	if !fromServer {
		s.packets <- pk
	}

	return true
}

func (s *httpServer) formatServerEvent(event string, pk packet.Packet) (string, error) {
	s.buf.Reset()
	s.sb.Reset()

	//fmt.Println(pk)
	//if uk, ok := pk.(packet.Unknown); ok {
	//fmt.Println("ul payload", string(uk.Payload))
	//}

	val := reflect.ValueOf(pk)
	var name string
	switch val.Kind() {
	case reflect.Struct:
		name = val.Type().Name()

	case reflect.Pointer:
		name = val.Elem().Type().Name()
	}
	m := map[string]any{
		"data": map[string]any{
			"id":     pk.ID(),
			"name":   name,
			"struct": pk,
		}}

	err := s.enc.Encode(m)
	if err != nil {
		return "", err
	}

	s.sb.WriteString(fmt.Sprintf("event: %s\n", event))
	s.sb.WriteString(fmt.Sprintf("data: %v\n\n", s.buf.String()))

	return s.sb.String(), nil
}
