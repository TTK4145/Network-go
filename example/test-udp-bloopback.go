package main 


import (
	"fmt"
	"Network-go/udp"
	"time"
)

func print_udp_message(msg udp.Udp_message){
	fmt.Printf("msg:  \n \t raddr = %s \n \t data = %s \n \t length = %v \n", msg.Raddr, msg.Data, msg.Length)
}

func node (send_ch, receive_ch chan udp.Udp_message){
for {
	time.Sleep(1*time.Second)
	snd_msg := udp.Udp_message{Raddr:"broadcast", Data:"Hello World", Length:11}
	fmt.Printf("Sending------\n")
	send_ch <- snd_msg
	print_udp_message(snd_msg)
	fmt.Printf("Receiving----\n")
	rcv_msg:= <- receive_ch
	print_udp_message(rcv_msg)
	}
}


func main (){
	send_ch := make (chan udp.Udp_message)
	receive_ch := make (chan udp.Udp_message)
	err := udp.Udp_init(20001, 20002, 1024, send_ch, receive_ch)	
	go node(send_ch, receive_ch)


	if (err != nil){
		fmt.Print("main done. err = %s \n", err)
	}
		neverReturn := make (chan int)
	<-neverReturn

}
