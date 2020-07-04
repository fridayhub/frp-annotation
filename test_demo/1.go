package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	service := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go read(conn)
		go read(conn)
		//daytime := time.Now().String()
		//conn.Write([]byte(daytime)) // don't care about return value
		//conn.Close()                // we're finished with this client
	}
}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func read(conn net.Conn) {

	var data []byte
	_, err := conn.Read(data)
	if err != nil {
		fmt.Printf("read error:%v\n", err)
	}
	fmt.Printf("read data:%v\n", data)

}
