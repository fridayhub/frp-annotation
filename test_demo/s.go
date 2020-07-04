package main

// 多路复用
import (
	"fmt"
	"github.com/hashicorp/yamux"
	"net"
	"time"
)

func recv(s net.Conn, id int) {
	for {
		buf := make([]byte, 4)
		n, err := s.Read(buf)
		if err == nil {
			fmt.Println("r ID:", id, ", r len:", n, time.Now().Unix(), string(buf))
		} else {
			fmt.Println(time.Now().Unix(), err)
			return
		}
	}
}

func Recv(stream net.Conn, id int) {

	buf := make([]byte, 4)
	n, err := stream.Read(buf)
	go recv(stream, id)
	if err == nil {
		fmt.Println("ID:", id, ", len:", n, time.Now().Unix(), string(buf))
	} else {
		fmt.Println(time.Now().Unix(), err)
		return
	}
}
func main() {
	// 建立底层复用连接
	tcpaddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8980")
	tcplisten, _ := net.ListenTCP("tcp", tcpaddr)
	conn, _ := tcplisten.Accept()
	session, _ := yamux.Server(conn, nil)

	id := 1
	for {
		// 建立多个流通路
		stream, err := session.Accept()
		if err == nil {
			fmt.Println("accept")
			id++
			go Recv(stream, id)
		} else {
			fmt.Println("session over.")
			return
		}
	}
}
