package main

import (
	"log/slog"
	"net/http"
)

func main() {
	// create the HTTP client
	c := &http.Client{}

	// make a GET HTTP request root of the server
	res, err := c.Get("http://localhost:8080")
	if err != nil {
		slog.Error("failed to send the request", "err", err)

		return
	}

	slog.Info("got the response", "status", res.Status)
}

// func main() {
// 	// create the HTTP client
// 	c := &http.Client{}
//
// 	// make a GET HTTP request root of the server
// 	res, err := c.Post("http://localhost:8080", "text/plain", strings.NewReader("Gopher!!"))
// 	if err != nil {
// 		slog.Error("failed to send the request", "err", err)
//
// 		return
// 	}
//
// 	slog.Info("got the response", "status", res.Status)
// }

// func main() {
// 	// create the HTTP client
// 	c := &http.Client{}
//
// 	// make a GET HTTP request root of the server
// 	res, err := c.Post("http://localhost:8080", "text/plain", strings.NewReader("Gopher!!"))
//
// 	// see we are ignoring the EOF error here
// 	if err != nil && !errors.Is(err, io.EOF) {
// 		slog.Error("failed to send the request", "err", err)
//
// 		return
// 	}
//
// 	// check if the status code is 2xx
// 	if res.StatusCode != http.StatusOK {
// 		slog.Error("wrong status", "status", res.Status)
// 		return
// 	}
//
// 	// print the status code
// 	slog.Info("status code", "code", res.StatusCode)
//
// 	// read the body
// 	body, _ := io.ReadAll(res.Body)
//
// 	// close the body before exiting
// 	defer res.Body.Close()
//
// 	// print the body
// 	slog.Info("response body", "body", body)
// }

// func main() {
// 	// create the HTTP client
// 	c := &http.Client{}
//
// 	// make a GET HTTP request root of the server
// 	res, err := c.Get("http://localhost:8080")
//
// 	// see we are ignoring the EOF error here
// 	if err != nil && !errors.Is(err, io.EOF) {
// 		slog.Error("failed to send the request", "err", err)
//
// 		return
// 	}
//
// 	// check if the status code is 2xx
// 	if res.StatusCode != http.StatusOK {
// 		slog.Error("wrong status", "status", res.Status)
// 		return
// 	}
//
// 	// print the status code
// 	slog.Info("status code", "code", res.StatusCode)
//
// 	// read the body
// 	body, _ := io.ReadAll(res.Body)
//
// 	// close the body before exiting
// 	defer res.Body.Close()
//
// 	// print the body
// 	slog.Info("response body", "body", body)
// }
