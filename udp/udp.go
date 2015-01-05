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

type udp_connection struct {
	stuff string


}

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
/*
import (
	"fmt"
	"net"
	"strconv"
	"errors"
)

type Udp_registration struct {
	Function			string
	IP						string
	Port 					int
	Receive_ch		chan []byte
	Send_ch				chan []byte
//	Data_ch				chan []byte
	Message_size	int
}

type udp_if	struct {
	function			string
	conn 					*net.UDPConn
	addr 					*net.UDPAddr
	message_size	int
	receive_ch 		chan []byte
	send_ch				chan []byte
//	data_ch				chan []byte
	index					int
	open					bool
	id						int
}

var udp_if_id_counter int
var udpLocal *net.UDPAddr
var udpBroadcast *net.UDPAddr

func debug_print_udp_if(i *udp_if) {
	fmt.Printf("UDP_IF: \n \t function: %s \n \t conn: %x \n \t addr: %x \n \t \t addr.string: %v \n \t \t addr.network: %s \n \t message_size: %v \n \t data_ch: %x \n \t index: %v \n \t open: %v \n \t id: %v \n \t \n", i.function, i.conn, i.addr, i.addr.String(), i.addr.Network(), i.message_size, i.data_ch, i.index, i.open, i.id)
}


//array to keep track of registered udp connection, used for closing
var udp_register_array [](*udp_if) 

func udp_add_if_register_array(udpConnection *udp_if) (err error){
	//add the new connection to the register array
	for i:= 0; i<len(udp_register_array); i++ {
		if udp_register_array[i].open == false {
			udp_register_array[i] = udpConnection
			udpConnection.index = i
			return nil
		}
	}
	udpConnection.index = len(udp_register_array)
	udp_register_array = append (udp_register_array, udpConnection)
	return nil
}


func Udp_init() (err error) {
	udp_if_id_counter = 0
	udp_register_array = make ([](*udp_if),0)
	tempUDPAddr, err := net.ResolveUDPAddr("udp4","255.255.255.255:21000")
	if err != nil { return err }
	tempConn, err := net.DialUDP("udp4",nil, tempUDPAddr)
	if err != nil { return err }
	addr := tempConn.LocalAddr()
//	fmt.Printf("Generating localaddress: Network(): %s \t String(): %s \n ", addr.Network(), addr.String()) 
	udpLocal, err = net.ResolveUDPAddr("udp4", addr.String())
	if err != nil { return err }
	udpLocal.Port = -1
	
	udpBroadcast, err = net.ResolveUDPAddr("udp4", "255.255.255.255:0")
	if err != nil { return err }
	udpBroadcast.Port = -1
	
	
	err = tempConn.Close()
	if err != nil { return err }
		
	return err
}

func udp_close_udp_if(udpConnection *udp_if) {
	if err := udpConnection.conn.Close(); err != nil {
		fmt.Printf("Error in udp_close_udp_if\n \t Closing the connection produced error: %s\n",err)
	}else {
		udpConnection.open = false			
	}
}

func udp_read_server(udpConnection *udp_if){
  defer func() {
        if r := recover(); r != nil {
            fmt.Println("ERROR in udp_read_server: %s \n Closing connection.", r)
						udp_close_udp_if(udpConnection)            
        }
    }()
	fmt.Printf("udp_read_server initialized\n")
	debug_print_udp_if(udpConnection)
	
	buf := make ([]byte,udpConnection.message_size)
	for {	
		n , err := udpConnection.conn.Read(buf)
//		fmt.Printf("Server Read \"%s\" \n", string(buf))
		udpConnection.receive_ch <- buf
		if (err != nil || n < 0 ) {
			fmt.Printf("udp_read_server ERROR::\n")
			panic(err)
		}	
	}			
}

func udp_write_server (udpConnection *udp_if){
  defer func() {
        if r := recover(); r != nil {
            fmt.Println("ERROR in udp_write_server: %s \n Closing connection.", r)
						udp_close_udp_if(udpConnection)            
        }
    }()
	fmt.Printf("udp_write_server initialized\n")
	
	debug_print_udp_if(udpConnection)
	
	for {
				buf := <- udpConnection.send_ch
//				fmt.Printf("Writing %s \n", str)
				n, err := udpConnection.conn.Write([]byte(buf))
				if (err != nil || n < 0 ) {
					fmt.Printf("udp_write_server ERROR::\n")
					panic(err)
				}
			}
}

func udp_create_listener(udpConnection *udp_if) (err error){
	fmt.Printf("Created Receiver\n");
	//create a connection
	udpConnection.conn, err = net.ListenUDP("udp4", udpConnection.addr)
	if err != nil || udpConnection.conn == nil {
	//	fmt.Printf("udp_create_listener:: ERROR %s %p \n", err, udpConnection.conn)
		return err
	} else {
		//create a listening server
	 err = udp_add_if_register_array(udpConnection)
	 if err != nil { return err }
	 go udp_read_server(udpConnection)
	}
	return err
}



func udp_create_transmitter(udpConnection *udp_if)(err error){
	fmt.Printf("Created Transmitter\n");

	//create a connection
	udpConnection.conn, err = net.DialUDP("udp4", nil ,udpConnection.addr)
	if err != nil || udpConnection.conn == nil {
//		fmt.Printf("udp_create_listener:: ERROR %s %p \n", err, udpConnection.conn)
		return err
	} else {
		//create a transmit server
		err = udp_add_if_register_array(udpConnection)
		if err != nil { return err }
		go udp_write_server(udpConnection)
	}
	return err
}

func udp_create_transreceiver(udpConnection *udp_if)(err error){
	fmt.Printf("Created Transmitter/Receiver\n");

	//create a connection
	udpConnection.conn, err = net.DialUDP("udp4", nil ,udpConnection.addr)
	if err != nil || udpConnection.conn == nil {
		fmt.Printf("udp_create_listener:: ERROR %s %p \n", err, udpConnection.conn)
		return err
	} else {
		//create a transmit server
		err = udp_add_if_register_array(udpConnection)
		if err != nil { return err }
//		go udp_write_server(udpConnection)
			go udp_read_server(udpConnection)
	}
	return err
}



func Udp_register(newRegistration Udp_registration) (id int, err error) {
	fmt.Printf("Udp_register::new registration\n")
	var udpConnection udp_if

	udpConnection.function = newRegistration.Function
	udpConnection.id = udp_if_id_counter
	udpConnection.message_size = newRegistration.Message_size
	udpConnection.open = true
		
	//generate a net.UDPAddr
	if newRegistration.IP == "local" {
		udpConnection.addr = &net.UDPAddr{udpLocal.IP, newRegistration.Port, ""}
	}else if newRegistration.IP == "broadcast" {
		udpConnection.addr = &net.UDPAddr{udpBroadcast.IP, newRegistration.Port, ""}
	} else {
		udpConnection.addr, err = net.ResolveUDPAddr("udp4", newRegistration.IP+":"+strconv.Itoa(newRegistration.Port))
		if err!=nil { return -1,err }
	}
	
	switch (newRegistration.Function){
		case "listener":
//			udpConnection.data_ch = newRegistration.Data_ch
			udpConnection.receive_ch = newRegistration.Receive_ch
			udpConnection.send_ch = nil
			err = udp_create_listener(&udpConnection)
		case "connection":
//			udpConnection.data_ch = newRegistration.Data_ch
			udpConnection.receive_ch = newRegistration.Receive_ch
			udpConnection.send_ch = newRegistration.Send_ch
			err = udp_create_transreceiver(&udpConnection)
//		case "transmit":
//			udpConnection.receive_ch = nil
//			udpConnection.data_ch = newRegistration.Data_ch
//			err = udp_create_transmitter(&udpConnection)
//		case "transmit/receive", "receive/transmit":
//			udpConnection.receive_ch = newRegistration.Receive_ch
//			udpConnection.send_ch = newRegistration.Send_ch
//			err = udp_create_transreceiver(&udpConnection)
		default:
			fmt.Printf("Error in Udp_register: function element is incorrectly formatted \n Please use either of the following formats: \n \t receive \n \t transmit \n\t transmit/receive \n\t receive/transmit \n")
			err= errors.New("Udp_register: incorrect function assignment")
			return -1, err
	}
	udp_if_id_counter++
	return id,err
}



func Udp_close( id int) (err error) {
	for i := 0; i <len(udp_register_array); i++ {
		if udp_register_array[i].id == id {
			if err = udp_register_array[i].conn.Close(); err != nil {
				return err
			}else{
				udp_register_array[i].open = false
				break				
			}
		}	
	}
	Udp_cleanup()
	return nil
}

func Udp_cleanup(){
	temp := make ([](*udp_if), 0)
	for  i := 0; i< len(udp_register_array); i++ {
		if udp_register_array[i].open == true {
			temp = append(temp, udp_register_array[i])		
		}	
	}
	udp_register_array = temp
}

func Udp_close_all() (err error){
	for i := 0; i<len(udp_register_array); i++ {
		if err =	udp_register_array[i].conn.Close(); err != nil {
			return err			
		}else {
			udp_register_array[i].open = false
		}
	}
	Udp_cleanup()
	if len(udp_register_array) != 0 {
		err = errors.New("Udp_close_all: Failed to close all connections")
		return err
	}
	return nil	
}

*/
