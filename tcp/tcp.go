package tcp


import (
	"fmt"
	"net"
	"strconv"
	"time"
	"errors"
	)
	
type Tcp_message struct{
	Raddr string
	Data string 
	Length int
}
	
var conn_list map[string]*net.TCPConn
	
type tcp_conn struct {
	conn *net.TCPConn
	receive_ch chan Tcp_message

}

func Tcp_init(localListenPort int, send_ch, receive_ch chan Tcp_message) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR recovered in tcp_init: %s ", r)
			}
	}()
	fmt.Println("Initializing TCP")
	
	conn_list = make(map[string]*net.TCPConn)
	
	baddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(20323))
	if err != nil {
		fmt.Println("Could not resolve baddr")
		return err
	}

	
	tempConn, err := net.DialUDP("udp4", nil, baddr)
	if err != nil {
		fmt.Println("Failed to dial baddr for laddr generation")
		return err
	}
	tempAddr := tempConn.LocalAddr()
	laddr, err := net.ResolveTCPAddr("tcp4", tempAddr.String())
	if err != nil {
		fmt.Println("Failed to resolve laddr")
		return err
	}
	laddr.Port = localListenPort
	
	tempConn.Close()
	

	listener, err := net.ListenTCP("tcp4",laddr)

	if err != nil {
		fmt.Println("Failed to initialize listener")
		return err
	}
	
	go tcp_transmit_server(send_ch)
	
	for {
		newConn, err := listener.AcceptTCP()
		fmt.Println("Received new request for connection")
		
		
		
		
		if err != nil {
			fmt.Printf("Error: accepting tcp conn \n")
			panic(err)
		}
		raddr := newConn.RemoteAddr()
		conn_list[raddr.String()] = newConn
		
		//setting up a reading server on each new connection
		go func (raddr string, conn *net.TCPConn, receive_ch chan Tcp_message){ 
			fmt.Println("Starting new tcp read server")
			buf := make([]byte,1024)
			for {
					n, err :=	conn.Read(buf)
					if err != nil || n < 0 {
						fmt.Printf("Error: tcp reader\n")
						panic(err)
					}
					receive_ch <- Tcp_message{Raddr: raddr, Data: string(buf), Length: n}
			}		
		}(raddr.String(), newConn, receive_ch)
		
	}
}

func tcp_transmit_server (ch chan Tcp_message){
	for {
		msg := <- ch
		fmt.Println("New message to send")
		
		_ , ok := conn_list[msg.Raddr]
		if (ok != true ){
			new_tcp_conn(msg.Raddr)//dial new tcp4		
		}
		
		sendConn, ok  := conn_list[msg.Raddr]
		if (ok != true) {
			err := errors.New("Failed to add newConn to map")
			panic(err)
		}
		
		n, err := sendConn.Write([]byte(msg.Data))	
		if err != nil || n < 0 {
			fmt.Printf("Error: tcp transmit server \n")
			panic(err)
		}
	}
}


func new_tcp_conn(raddr string) bool{
	fmt.Println("Adding new conn to list")
	//create address
	addr, err := net.ResolveTCPAddr("tcp4", raddr)
	if err != nil {
		fmt.Println("ERROR: new tcp conn, could not resolve addr")
		return false
	}

	for {
		newConn, err := net.DialTCP("tcp4", nil,  addr)
		
		if err != nil {
			fmt.Println("DialTCP failed, raddr : %s", raddr)
				time.Sleep(500*time.Millisecond)
		} else {
			conn_list[raddr] = newConn
			return true//got it BREAK!
		}
	}
}
