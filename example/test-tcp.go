package main 

import (
	"fmt"
	"../tcp"
	"time"
)


func main (){
	fmt.Println("Starting test-tcp.go")
	lport := 20234
	rchan := make (chan tcp.Tcp_message)
	schan := make (chan tcp.Tcp_message)
	tcp.Tcp_init(lport, schan, rchan)
	
	go func(ch chan tcp.Tcp_message){
		id := 0
		msg := {Raddr: "129.241.187.XXX:20233", Data := strconv.Itoa(id), Length = 32}

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
