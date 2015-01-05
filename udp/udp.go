package udp

import (
	"fmt"
	"net"
	"strconv"
)

var Global_Send_ch chan Udp_message
var Global_Receive_ch chan Udp_message

var laddr *net.UDPAddr //Local address 
var baddr *net.UDPAddr //Broadcast address

type Udp_message struct {

	raddr	string  //if receiving raddr=senders address, if sending raddr should be set to "broadcast" or an ip:port
	data string //TODO: implement another encoding, strings are meh
	length int //length of received data, in #bytes
}


func Udp_init (localListenPort, broadcastListenPort, message_size int) (err error) {
	//Creating local listening connections
	localListenConn, err := net.ListenUDP("udp4", nil)
	if err != nil { return err }
	
	//Generating local address
	tempAddr := localListenConn.LocalAddr()
	laddr, err = net.ResolveUDPAddr("udp4",tempAddr.String())
	
	//Generating broadcast address
	baddr, err = net.ResolveUDPAddr("udp4","255.255.255.255:"+strconv.Itoa(broadcastListenPort))
	if err != nil { return err }
	
	//Creating listener on broadcast connection
	broadcastListenConn , err := net.ListenUDP("udp", baddr)
	if err != nil { 
		localListenConn.Close()
		return err 
	}
	
	go udp_receive_server(localListenConn, broadcastListenConn, message_size)
	go udp_transmit_server(localListenConn, broadcastListenConn)
	
	fmt.Printf("Generating localaddress: Network(): %s \t String(): %s \n ", laddr.Network(), laddr.String()) 
	fmt.Printf("Generating broadcastaddress: Network(): %s \t String(): %s \n ", baddr.Network(), baddr.String()) 	
	return err
}


//0,5 points to the group who first sends me an email explaining why i dont need to synchronize the reading/tranmitting on the udp-connection
func udp_transmit_server (lconn, bconn *net.UDPConn){
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("ERROR in udp_transmit_server: %s \n Closing connection.", r)
			lconn.Close(); bconn.Close()
        }
    }()		
    var err error 
    var n int
	
	for {
		buf := <- Global_Send_ch
//				fmt.Printf("Writing %s \n", string(buf))
		if buf.raddr == "broadcast" {
			n, err = lconn.WriteToUDP([]byte(buf.data), baddr)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", buf.raddr)
			if (err != nil) {
				fmt.Printf("Error: udp_transmit_server: could not resolve raddr\n")
				panic(err)
			}
			n, err = lconn.WriteToUDP([]byte(buf.data), raddr)
		}
		if (err != nil || n < 0) {
			fmt.Printf("Error: udp_transmit_server: writing\n")
			panic(err)
		}
	}
}


func udp_receive_server (lconn, bconn *net.UDPConn, message_size int){
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("ERROR in udp_receive_server: %s \n Closing connection.", r)
			lconn.Close(); bconn.Close()
        }
    }()		 
    
	bconn_rcv_ch := make (chan Udp_message)
 	lconn_rcv_ch := make (chan Udp_message)
	
	go udp_connection_reader(lconn, message_size, lconn_rcv_ch)
	go udp_connection_reader(bconn, message_size, bconn_rcv_ch)

	for {
		select {
		
			case buf := <- bconn_rcv_ch:
				Global_Receive_ch <- buf
			
			case buf := <-lconn_rcv_ch:
				Global_Receive_ch <- buf
		}
	}
}


func udp_connection_reader(conn *net.UDPConn, message_size int, rcv_ch chan Udp_message){
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("ERROR in udp_connection_reader: %s \n Closing connection.", r)
			conn.Close()
        }
    }()	
    
	buf := make ([]byte, message_size)
		
	for {
		n, raddr, err := conn.ReadFromUDP(buf)
		if (err != nil || n < 0 ) {
			fmt.Printf("Error: udp_connection_reader: reading\n")
			panic(err)
		}
		rcv_ch <- Udp_message{raddr.String(), string(buf), n}
	}
}
