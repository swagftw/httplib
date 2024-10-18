package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var connChan = make(chan net.Conn)

func main() {
	// to listen to http traffic
	// we need to listen to tcp file descriptor
	// because at the end of the day http traffic uses tcp for transmission
	// so we create a tcp listener or in os words open a file descriptor

	// go provides you with net library which is nothing but the language interface to
	// IO, TCP/IP, UDP etc

	// we create a listener by providing the protocol TCP in this case and the host:port
	// which returns our listener if it is there or an error,
	// error can be returned in case of port is already in use (commonly this is the issue).
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		slog.Error("failed to create the listener", "err", err)

		return
	}

	slog.Info("started the listener, hit CTRL+C to quit")

	// create the OS signal channel to capture interrupt and terminate signals
	signalChan := make(chan os.Signal)

	// hook the channel to be notified for the signal
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// we will talk about this later
	go handleConnections(connChan)

	// a single listener can listen to stream of data over multiple connections
	// connection handling depends on the kind of protocol you are trying to use/make
	// so if multiple clients are trying to open connections to our listener we need to loop over the listener

	// but we have to run this loop inside a go routine so that we can gracefully shut down the application
	go func() {
		for {
			// net/conn is an interface to the actual connection object,
			// you can pass around the connection without pointers just like slices and interfaces.
			var conn net.Conn

			conn, err = listener.Accept()
			if err != nil {
				// if the error is listener closed, we just log it and close the go routine
				// if we don't do this the loop will continue to iterate over closed listener,
				// until the main go routine is completely shutdown
				if errors.Is(err, net.ErrClosed) {
					slog.Info("listener closed")

					return
				}

				slog.Error("failed to accept the connection", "err", err)

				// if error just move to the next connection
				continue
			}

			// send the connection to the channel to listen for data and magical operations
			// this way we can unblock the loop and not care about how connection is getting handled really
			connChan <- conn
		}
	}()

	// this select block will block the main goroutine till the interrupt signal is received
	select {
	case <-signalChan:
		// graceful shutdown, will close the connection channel and tcp listener
		close(connChan)
		_ = listener.Close()
		slog.Info("Bubye fellow gophers!")
	}
}

// handleConnections will handle the connections for us
func handleConnections(connChan chan net.Conn) {
	// just slices and maps, channels in go are also iterable,
	// and they will iterate till the channel is closed
	for conn := range connChan {
		go handleConnection_v1(conn)
	}
}

// handleConnection is the worst implementation of what we really want to achieve,
// but it's a good start
func handleConnection_v1(conn net.Conn) {
	// if we look at the methods available on the net.Conn,
	// which interests us for now is Read method,

	// conn.Read takes a slice of byte as an argument, so let's create a byte slice and pass to it
	// conn.Read reads the data from connection and writes it to the passed slice of byte
	// we have to explicitly say the size of the byte slice otherwise reader will not read data into it
	dataSlice := make([]byte, 1<<10)

	// read the data in the dataSlice
	_, err := conn.Read(dataSlice)
	// check for the error
	if err != nil {
		// if the error is NOT because of EOF error we log it and return from whole function
		if !errors.Is(err, io.EOF) {
			slog.Error("failed to read the data from connection", "err", err)

			return
		}
	}

	// for now just log the data that we read
	// and try to understand what this binary data really is
	// for that we have to understand binary and non-binary data types
	// for example: what is the difference between binary and non-binary strings?
	slog.Info("read data", "data", dataSlice)

	// we can see there is some structure to this data and let's try to understand this structure
	// GET / HTTP/1.1\r\n                   ----- This is the first line (Notice that it ends with \r\n)
	// Host: localhost:8080\r\n             ----- This is the second line (Notice that it ends with \r\n)
	// User-Agent: Go-http-client/1.1\r\n   ----- This is the third line (Notice that it ends with \r\n)
	// Accept-Encoding: gzip\r\n\r\n        ----- This is the fourth line (Notice that it end with A DOUBLE \r\n LOL)

	// let's try to figure out the meaning of this gibberish looking data
	// Keeping already known things in the mind

	// FIRST LINE -- GET is method -- / is path -- HTTP/1.1 is protocol version -- \r\n is CRLF |--- (protocol info)

	// SECOND LINE -- Host: host with port -- \r\n is CRLF              |
	// THIRD LINE -- User-Agent: with value of agent -- \r\n is CRLF    |--- (Headers)
	// FOURTH LINE -- data encoding which ends with double \r\n         |

	// now that double CRLF can also be interpreted as new line and an empty line
	// which can be used to understand SOMETHING!!
	// and that SOMETHING is nothing but DATA
	// but where's the data now? Let's send some data then

	// POST / HTTP/1.1\r\n
	// Host: localhost:8080\r\n
	// User-Agent: Go-http-client/1.1\r\n
	// Content-Length: 8\r\n
	// Content-Type: text/plain\r\n
	// Accept-Encoding: gzip\r\n
	// \r\n -- NOTICE this empty line which divides headers from data
	// Gophers!! -- NOTICE there is no CRLF at the end, why?

	// but how do we know how much data client is going to send us?
}

func handleConnection_v2(conn net.Conn) {
	// another way of dealing with the unknown amount of data is to
	// read the data in small chunks and stop if we read data less than size of byte slice
	// and keep on adding the data to another large byte slice

	completeData := make([]byte, 0, 1<<10)

	for {
		dataSlice := make([]byte, 1<<6)
		n, err := conn.Read(dataSlice)
		if err != nil {
			// if the error is NOT because of EOF error we log it and return from whole function
			if !errors.Is(err, io.EOF) {
				slog.Error("failed to read the data from connection", "err", err)

				return
			}
		}

		// add the read data into larger byte slice
		completeData = append(completeData, dataSlice...)

		if n < 1<<6 {
			break
		}
	}

	slog.Info("data read", "data", completeData)

	// after reading all the data we close the connection :)
	// forgive me error, I am ignoring you for now
	_ = conn.Close()
}

func handleConnection_v3(conn net.Conn) {
	// what magical are we going to do here?

	// we are trying to implement an HTTP/1.1 protocol here,
	// and RFC states that server should close the connection after server sends response back to the client
	// but as you see if we close the connection from server side, client cries with EOF
	// lucky us, EOF can be handled the way we want.

	// now about the magical part, lets try to send some data to the client from our connection

	completeData := make([]byte, 0, 1<<10)

	for {
		dataSlice := make([]byte, 1<<6)
		n, err := conn.Read(dataSlice)
		if err != nil {
			// if the error is NOT because of EOF error we log it and return from whole function
			if !errors.Is(err, io.EOF) {
				slog.Error("failed to read the data from connection", "err", err)

				return
			}
		}

		// add the read data into larger byte slice
		completeData = append(completeData, dataSlice...)

		if n < 1<<6 {
			break
		}
	}

	// now lets just send some random data on the connection
	// to do that we need to use conn.Write method which will let us write on the connection
	// and TCP protocol will transfer that to client over internet/ethernet in this case

	// for now, we are not trying to make sense of any data sent by the client and
	// will write some random bytes on the connection to see how client will react to it

	_, err := conn.Write([]byte("Gopher!!"))
	if err != nil {
		slog.Error("failed to write to the connection", "err", err)

		return
	}

	// and now we close the connection
	_ = conn.Close()

	// we see the client without error now, we skipped the right error but the response is still nil
	// because?
}

func handleConnection_v4(conn net.Conn) {
	// let's try to write the proper response this time,
	// but we still have a dumb version of reading the data which is fine for now
	completeData := make([]byte, 0, 1<<10)

	for {
		dataSlice := make([]byte, 1<<6)
		n, err := conn.Read(dataSlice)
		if err != nil {
			// if the error is NOT because of EOF error we log it and return from whole function
			if !errors.Is(err, io.EOF) {
				slog.Error("failed to read the data from connection", "err", err)

				return
			}
		}

		// add the read data into larger byte slice
		completeData = append(completeData, dataSlice...)

		if n < 1<<6 {
			break
		}
	}

	// now rather than writing some stupid data over the connection lets write something that
	// http client will understand and will not give us empty response
	buffer := bytes.Buffer{}

	dataToWrite := []byte("Gopher!!")

	buffer.WriteString("HTTP/1.1 200 OK")              // protocol and status don
	buffer.WriteString("\r\n")                         // CRLF
	buffer.WriteString("Content-Length: ")             // header for content length
	buffer.WriteString(strconv.Itoa(len(dataToWrite))) // actual length of the data
	buffer.WriteString("\r\n")                         // CRLF
	buffer.WriteString("\r\n")                         // empty line for data
	buffer.Write(dataToWrite)                          // actual data

	// write the buffer we created on the connection
	_, err := conn.Write(buffer.Bytes())
	if err != nil {
		slog.Error("failed to write on the connection", "err", err)

		return
	}

	// and now close the connection
	_ = conn.Close()
}

func handleConnection_v5(conn net.Conn) {
	// this time we will read the data from connection in a better way
	// to do that we can leverage the pattern we understood from request

	// this is a special scanner, which lets you read the data from connection reader
	// as a line which we can use to split our request
	scanner := bufio.NewScanner(conn)

	req := new(Request)
	req.Headers = make(map[string]string)

	// scan will iterate over the connection data and read the line and store it in buffer
	for scanner.Scan() {
		// that buffer we can access via scanner.Bytes()
		line := scanner.Text()

		err := req.parse(line)
		if err != nil {
			// if error is there while parsing request written error
			_ = conn.Close()

			return
		}

		// check if the request reading is done
		if req.done {
			break
		}
	}

	// now we have the request
	// now rather than writing some stupid data over the connection lets write something that
	// http client will understand
	buffer := bytes.Buffer{}

	// a little bit of router like behaviour
	if req.Path != "/" || req.Method != "GET" {
		buffer.WriteString("HTTP/1.1 404 Not Found")
		buffer.WriteString("\r\n")
		buffer.WriteString("Content-Length: 0")
		buffer.WriteString("\r\n")
		buffer.WriteString("\r\n")

		_, _ = conn.Write(buffer.Bytes())

		_ = conn.Close()

		return
	}

	dataToWrite := []byte("Gopher!!")

	buffer.WriteString("HTTP/1.1 200 OK")              // protocol and status don
	buffer.WriteString("\r\n")                         // CRLF
	buffer.WriteString("Content-Length: ")             // header for content length
	buffer.WriteString(strconv.Itoa(len(dataToWrite))) // actual length of the data
	buffer.WriteString("\r\n")                         // CRLF
	buffer.WriteString("\r\n")                         // empty line for data
	buffer.Write(dataToWrite)                          // actual data

	// write the buffer we created on the connection
	_, err := conn.Write(buffer.Bytes())
	if err != nil {
		slog.Error("failed to write on the connection", "err", err)

		return
	}

	// and now close the connection
	_ = conn.Close()
}

type Request struct {
	Protocol      string
	Method        string
	Path          string
	Headers       map[string]string
	wroteProtocol bool
	wroteHeaders  bool
	Body          []byte
	done          bool
}

// parseProtocol will parse the protocol
func (r *Request) parse(line string) error {
	// if the line is CRLF headers are over
	if len(line) == 0 {
		r.wroteHeaders = true
		r.done = true // ideally should be done after reading the body if any
		return nil
	}

	// if protocol is not written parse the protocol line
	if !r.wroteProtocol {
		splits := strings.Split(line, " ")

		if len(splits) != 3 {
			return fmt.Errorf("protocol line is not valid")
		}

		r.Method = splits[0]
		r.Path = splits[1]
		r.Protocol = splits[2]

		r.wroteProtocol = true

		return nil
	}

	// if headers are not written, add headers
	if !r.wroteHeaders {
		splits := strings.SplitN(line, ": ", 2)

		if len(splits) != 2 {
			return fmt.Errorf("invalid header")
		}

		r.Headers[strings.ToLower(splits[0])] = strings.ToLower(splits[1])

		return nil
	}

	// skip the body for now

	return nil
}
