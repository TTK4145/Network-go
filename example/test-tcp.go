//go run test-tcp.go -raddr="129.241.187.145:20033" -lport=20034
package main 

import (
	"fmt"
	"../tcp"
	"time"
	"strconv"
	"flag"
)



var raddr = flag.String("raddr", "127.0.0.1:21331", "the ip adress for the target connection")
var lport = flag.Int("lport", 21337, "the local port to listen on for new conns")


func main (){
	flag.Parse()
	fmt.Println("Starting test-tcp.go")
	rchan := make (chan tcp.Tcp_message)
	schan := make (chan tcp.Tcp_message)
	tcp.Tcp_init(*lport, schan, rchan)
	
	go func(ch chan tcp.Tcp_message){
		id := 0
		msg := tcp.Tcp_message{Raddr: *raddr, Data: strconv.Itoa(id), Length: 32}

		for {
			msg.Data = strconv.Itoa(id)
			schan <- msg
			id++
			fmt.Println("%v Sent: %v", lport,msg)
			time.Sleep(1*time.Second)	
		}	
	}(schan)

	for {
		msg := <- rchan
		fmt.Println("%v Received: %v",lport, msg)	
	}
}
