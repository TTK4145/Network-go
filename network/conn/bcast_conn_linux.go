// +build linux

package conn

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil { fmt.Println("Error: Socket:", err) }
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil { fmt.Println("Error: SetSockOpt REUSEADDR:", err) }
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil { fmt.Println("Error: SetSockOpt BROADCAST:", err) }
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil { fmt.Println("Error: Bind:", err) }

	f := os.NewFile(uintptr(s), "")
	conn, err := net.FilePacketConn(f)
	if err != nil { fmt.Println("Error: FilePacketConn:", err) }
	f.Close()

	return conn
}
