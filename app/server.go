package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	fmt.Println("Server listening on port 6379")
	table := make(map[string]string)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleClient(conn, table)
	}
}

func handleClient(conn net.Conn, table map[string]string) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if errors.Is(err, io.EOF) {
			fmt.Println("Client closed the connections:", conn.RemoteAddr())
			break
		} else if err != nil {
			fmt.Println("Error while reading the message")
		}

		buf = buf[:n]

		input, err := parseInput(buf)
		if err != nil {
			fmt.Println("Parse input error")
		}

		processInput(input, conn, table)
	}
}

func processInput(input []string, conn net.Conn, table map[string]string) {
	command := strings.ToUpper(input[0])

	switch command {
	case "PING":
		conn.Write([]byte("+PONG\r\n"))
	case "ECHO":
		if len(input) == 1 {
			conn.Write([]byte("+Echo command must include value\r\n"))
		} else {
			conn.Write([]byte(fmt.Sprintf("+%s\r\n", input[1])))
		}
	case "GET":
		if len(input) == 1 {
			conn.Write([]byte("+Get command must include key\r\n"))
		} else {
			if val, ok := table[input[1]]; ok {
				length := len(val)
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", length, val)))
			} else {
				conn.Write([]byte("+No value found\r\n"))
			}
		}
	case "SET":
		if len(input) != 3 {
			conn.Write([]byte("+Set command must include key and value\r\n"))
		} else {
			key := input[1]
			value := input[2]
			table[key] = value
			conn.Write([]byte(fmt.Sprintf("+SET Key: %s Value: %s\r\n", key, value)))
		}
	default:
		conn.Write([]byte(fmt.Sprintf("+Recieved unknown command: %s\r\n", command)))
	}
}

func parseInput(b []byte) ([]string, error) {
	i := 0

	if i == len(b) {
		return nil, io.ErrUnexpectedEOF
	}

	// Check for array prefix
	if b[i] != '*' {
		return nil, errors.New("array expected")
	}

	// Move to next character to get array length
	i++
	arrLen := getLength(b, &i)

	err := clrtExpected(b, &i)
	if err != nil {
		return nil, err
	}

	// Iterate through array
	var output []string
	for j := 0; j < arrLen; j++ {
		// BULK STRING
		if b[i] == '$' {
			i++
			stringLen := getLength(b, &i)
			clrtExpected(b, &i)
			var result string
			for k := 0; k < stringLen; k++ {
				result += string(b[i])
				i++
			}
			output = append(output, result)
			clrtExpected(b, &i)

		}
	}
	return output, nil
}

func getLength(b []byte, i *int) int {
	var length int
	for *i < len(b) && b[*i] >= '0' && b[*i] <= '9' {
		length = length*10 + int(b[*i]-'0')
		*i++
	}
	return length
}

func clrtExpected(b []byte, i *int) error {
	if *i+1 >= len(b) || b[*i] != '\r' || b[*i+1] != '\n' {
		return errors.New("CRLF expected")
	}
	*i += 2
	return nil
}
