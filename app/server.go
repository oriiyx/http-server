package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	// Uncomment this block to pass the first stage
	"net"
	"os"
)

const (
	IP = "4221"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	address := "0.0.0.0:" + IP
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			fmt.Println("Could not close the listener: ", err.Error())
			os.Exit(1)
		}
	}(l)

	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			HandleRequest(c)
			_ = c.Close()
		}(conn)
	}

}

func HandleRequest(conn net.Conn) {
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return
	}

	urlPath := request.URL.Path

	if request.Method == "GET" {

		if urlPath == "" || urlPath == "/" {
			HandleConnWriting(conn, "HTTP/1.1 200 OK", "", "")
			return
		}

		if strings.Contains(urlPath, "/echo/") {
			splitUrlPath := strings.Split(urlPath, "/")
			wildcard := splitUrlPath[len(splitUrlPath)-1]
			if wildcard == "" {
				HandleConnWriting(conn, "HTTP/1.1 404 Not Found", "", "")
			}
			wildcardLength := strconv.Itoa(len(wildcard))

			isCompressed := findIfAcceptsGzip(request)
			if isCompressed {
				compressionContent := compressString(wildcard)
				compressionLength := strconv.Itoa(len(compressionContent))
				HandleConnWriting(conn, "HTTP/1.1 200 OK", "Content-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: "+compressionLength+"\r\n", compressionContent)
			} else {
				HandleConnWriting(conn, "HTTP/1.1 200 OK", "Content-Type: text/plain\r\nContent-Length: "+wildcardLength+"\r\n", wildcard)
			}
		}

		if strings.Contains(urlPath, "/user-agent") {
			userAgent := request.Header.Get("User-Agent")
			userAgentLength := strconv.Itoa(len(userAgent))
			HandleConnWriting(conn, "HTTP/1.1 200 OK", "Content-Type: text/plain\r\nContent-Length: "+userAgentLength+"\r\n", userAgent)
		}

		if strings.Contains(urlPath, "/files/") {
			args := os.Args
			if len(args) > 2 {
				if args[1] == "--directory" {
					dir := os.Args[2]
					if _, err := os.Stat(dir); err == nil {
						splitUrlPath := strings.Split(urlPath, "/")
						wildcard := splitUrlPath[len(splitUrlPath)-1]
						data, err := os.ReadFile(dir + wildcard)
						if err != nil {
							HandleConnWriting(conn, "HTTP/1.1 404 Not Found", "", "")
						}
						dataLength := strconv.Itoa(len(data))
						HandleConnWriting(conn, "HTTP/1.1 200 OK", "Content-Type: application/octet-stream\r\nContent-Length: "+dataLength+"\r\n", string(data))
					} else {
						HandleConnWriting(conn, "HTTP/1.1 404 Not Found", "", "")
					}
				}
			}
		}

		HandleConnWriting(conn, "HTTP/1.1 404 Not Found", "", "")
	}

	if request.Method == "POST" {
		if strings.Contains(urlPath, "/files/") {
			args := os.Args
			if len(args) > 2 {
				if args[1] == "--directory" {
					dir := os.Args[2]
					splitUrlPath := strings.Split(urlPath, "/")
					wildcard := splitUrlPath[len(splitUrlPath)-1]
					_ = EnsureBaseDir(dir + wildcard)

					buf := make([]byte, 1024)
					_, _ = request.Body.Read(buf)
					r := strings.Split(string(buf), "\r\n")
					content := strings.Trim(r[len(r)-1], "\x00")

					writeErr := os.WriteFile(dir+wildcard, []byte(content), 0644)
					if writeErr != nil {
						fmt.Println(writeErr.Error())
					}

					HandleConnWriting(conn, "HTTP/1.1 201 Created", "", "")
				}
			}
		}
	}
}

func compressString(wildcard string) string {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	defer zw.Flush()

	_, err := zw.Write([]byte(wildcard))
	if err != nil {
		log.Fatal(err)
	}

	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}

	return buf.String()
}

func findIfAcceptsGzip(request *http.Request) bool {
	acceptEncoding := request.Header.Get("Accept-Encoding")

	if strings.Contains(acceptEncoding, "gzip") {
		return true
	}

	return false
}

func EnsureBaseDir(fpath string) error {
	baseDir := path.Dir(fpath)
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}

func HandleConnWriting(conn net.Conn, status, header, body string) {
	_, err := conn.Write([]byte(fmt.Sprintf("%s\r\n%s\r\n%s", status, header, body)))
	if err != nil {
		fmt.Println("Error writing back to connection: ", err.Error())
		os.Exit(1)
	}
}
