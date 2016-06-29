// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	// "time"
)

type multiEchoServer struct {
	ln                                                            net.Listener
	connectedClients                                              int
	clients                                                       map[net.Conn]chan string
	acceptedClients, deletedClient                                chan net.Conn
	messagesRead                                                  chan string
	clientsCountRequest, clientsCount, closeRequest, serverClosed chan int
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	m := &multiEchoServer{}
	// m.connectedClientsMutex = make(chan int, 1)
	m.clients = make(map[net.Conn]chan string) // initialize the clients map
	// m.connectedClientsMutex <- 1               // initialize the channel
	return m
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	p := strconv.Itoa(port)
	ln, err := net.Listen("tcp", ":"+p)
	if err != nil {
		fmt.Println("Couldn't listen: ", err)
		return err
	}
	mes.ln = ln
	mes.acceptedClients = make(chan net.Conn)
	mes.deletedClient = make(chan net.Conn)
	mes.messagesRead = make(chan string)
	mes.clientsCountRequest = make(chan int)
	mes.clientsCount = make(chan int)
	mes.closeRequest = make(chan int)
	mes.serverClosed = make(chan int)

	go handleAcc(mes)
	go dispatcher(mes)
	return nil
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	// fmt.Println("close")

	mes.closeRequest <- 1
	<-mes.serverClosed
	print("Close")
}

func (mes *multiEchoServer) Count() int {
	mes.clientsCountRequest <- 1
	count := <-mes.clientsCount
	return count
}

// each Accept spawns a new goroutine which continuously listens to clients
func handleAcc(mes *multiEchoServer) {
	for {
		// print("Acc")
		ln := mes.ln
		print("1")
		conn, err := ln.Accept() // accept is blocking
		if err != nil {
			fmt.Println("Couldn't accept: ", err)
			return
		}
		mes.acceptedClients <- conn
	}
}

func dispatcher(mes *multiEchoServer) {
	for {

		select {
		case conn := <-mes.acceptedClients:

			print("acceptedClients")
			// initialize the map for this particular conn with
			// the ClientMessages struct of channel with buffer 75
			mes.clients[conn] = make(chan string, 75)
			go handleConn(mes, conn)
			go handleClientWrites(mes, conn, mes.clients[conn])

		case msgRead := <-mes.messagesRead:
			print("messagesRead: " + msgRead)
			sendMessage(mes, msgRead)

		case <-mes.clientsCountRequest:
			print("clientsCountRequest")
			c := len(mes.clients)
			mes.clientsCount <- c

		case deleteConn := <-mes.deletedClient:
			delete(mes.clients, deleteConn)

		case <-mes.closeRequest:
			for c, _ := range mes.clients {
				c.Close()
			}
			mes.serverClosed <- 1

			// default:
			// if nothing happening, wait
			// time.Sleep(300 * time.Millisecond)
		}
	}
}

func handleConn(mes *multiEchoServer, conn net.Conn) {
	read := bufio.NewReader(conn) // create a reader to read messages from client
	for {
		// do error
		// print("Conn")
		msgB, err := read.ReadBytes('\n') // read messages delimited by byte
		if err != nil {
			mes.deletedClient <- conn
			return
		}
		l := len(msgB)
		// print("l = " + string(l))
		l--
		if l > 0 {
			msgB = msgB[:l]
			msg := string(msgB)
			msg = msg + msg + "\n"
			mes.messagesRead <- string(msg)
			// print("handleConn" + string(msg))
		}
	}
}

func handleClientWrites(mes *multiEchoServer, conn net.Conn, msgC chan string) {
	// do a range if it doesnt work
	// for {
	for msg := range msgC {
		// msg := <-msgC
		// print("handleClientWrites" + msg)
		_, err := conn.Write([]byte(msg))
		if err != nil {
			mes.deletedClient <- conn
			return
		}
	}
}

func sendMessage(mes *multiEchoServer, msg string) {
	for c, _ := range mes.clients {
		if len(mes.clients[c]) < 75 {
			mes.clients[c] <- msg
			// } else {
			// time.Sleep(100 * time.Millisecond)
		}
	}
}

func print(msg string) {
	// fmt.Println(msg)
}

// go rountine for read and write
// read forever reads
// write forever writes if there is in client channel
// when servers gets msg, makes the msgs, then puts in clients channel but check length
