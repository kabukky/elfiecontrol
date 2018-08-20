package main

import (
	"encoding/hex"
	"io"
	"net"
	"net/http"

	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/av/pubsub"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/format/flv"
	"github.com/nareix/joy4/format/rtmp"
)

func init() {
	format.RegisterAll()
}

var (
	videoStartData, _ = hex.DecodeString("000102030405060708092828")
)

type writeFlusher struct {
	httpflusher http.Flusher
	io.Writer
}

func (self writeFlusher) Flush() error {
	self.httpflusher.Flush()
	return nil
}

func getVideo() error {
	// Create connection
	conn, err := net.Dial("tcp", "172.16.10.1:8888")
	if err != nil {
		return err
	}
	_, err = conn.Write(videoStartData)
	if err != nil {
		return err
	}

	que := pubsub.NewQueue()

	cursor := que.Latest()
	avutil.CopyFile(rtmp.NewConn(conn), cursor)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/x-flv")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(200)
		flusher := w.(http.Flusher)
		flusher.Flush()

		muxer := flv.NewMuxerWriteFlusher(writeFlusher{httpflusher: flusher, Writer: w})
		cursor := que.Latest()
		avutil.CopyFile(muxer, cursor)
	})

	return http.ListenAndServe(":8089", nil)
}
