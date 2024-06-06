package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	fmt.Println("Server listening on port 6379")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
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

		fmt.Println("string(Buf[:n]): ", string(buf[:n]))
		buf = buf[:n]

		input, err := parseInput(buf)
		fmt.Println("String Arr: ", input)
		if err != nil {
			fmt.Println("Parse input error")
		}

		processInput(input, conn)
	}
}

func processInput(input []string, conn net.Conn) {
	command := input[0]

	switch command {
	case "PING":
		conn.Write([]byte("+PONG\r\n"))
	case "ECHO":
		if len(input) == 1 {
			conn.Write([]byte("+Must include ECHO value\r\n"))
		} else {
			conn.Write([]byte(fmt.Sprintf("+%s\r\n", input[1])))
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

	if b[i] != '*' {
		return nil, errors.New("array expected")
	}

	fmt.Println("Array found")

	i++

	var arrLen int

	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		arrLen = arrLen*10 + int(b[i]-'0')
		i++
	}

	fmt.Println("Total elements in array: ", arrLen)

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
			var stringLen int
			for i < len(b) && b[i] >= '0' && b[i] <= '9' {
				stringLen = stringLen*10 + int(b[i]-'0')
				i++
			}
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

func clrtExpected(b []byte, i *int) error {
	if *i+1 >= len(b) || b[*i] != '\r' || b[*i+1] != '\n' {
		return errors.New("CRLF expected")
	}
	*i += 2
	return nil

}
