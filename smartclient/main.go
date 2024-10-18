package main

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
)

func main() {
	// just like net.Listen we have net.Dial as well
	// which dials connection to the local/remote file descriptor
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		slog.Error("failed to dial the connection", "err", err)

		return
	}

	// TODO: use the buffer and write your own request
	// create the new request using http lib, because we don't know how to create a raw request yet (but we will soon!)
	r, err := http.NewRequest("GET", "http://localhost:9090", nil)
	if err != nil {
		slog.Error("failed to create request", "err", err)

		return
	}

	// get the dump of the request that will be sent on the wire
	requestData, _ := httputil.DumpRequest(r, false)

	// write that dump to the connection
	_, _ = conn.Write(requestData)

	// same problem here as well how much to read?
	// let's not worry about that yet
	respData := make([]byte, 1<<10)

	// read the response data
	_, err = conn.Read(respData)
	if err != nil {
		slog.Error("failed to read the data", "err", err)

		return
	}

	// and check out what is the format now
	slog.Info("read data", "data", respData)

	// this is the format
	// HTTP/1.1 200 OK\r\n --- protocol --- status --- CRLF
	// Headers\r\n -- headers -- CRLF
	// \r\n -- empty line
	// body
}
