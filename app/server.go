package main

import (
	"bufio"
	"fmt"
	"net/http"
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

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	HandleRequest(conn)
}

func HandleRequest(conn net.Conn) {
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return
	}

	urlPath := request.URL.Path

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
		HandleConnWriting(conn, "HTTP/1.1 200 OK", "Content-Type: text/plain\r\nContent-Length: "+wildcardLength+"\r\n", wildcard)
	}

	HandleConnWriting(conn, "HTTP/1.1 404 Not Found", "", "")
}

func HandleConnWriting(conn net.Conn, status, header, body string) {
	_, err := conn.Write([]byte(fmt.Sprintf("%s\r\n%s\r\n%s", status, header, body)))
	if err != nil {
		fmt.Println("Error writing back to connection: ", err.Error())
		os.Exit(1)
	}
}
