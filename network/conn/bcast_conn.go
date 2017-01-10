// +build !windows

package conn

import (
	"net"
	"os"
	"syscall"
)

/*
I have not found a way to create a proper broadcast socket on Windows (one that
doesn't give an "address already in use" error). This is not a problem with
Windows or sockets, but some quirk of the Go library. It is impossible to set
the required *broadcast* and *reuseaddr* socket options when using the
net.Dial/.Listen functions, so the only option is using the `syscall` functions.
But I can't turn the syscall.Socket into a os.File that can be passed to
net.FilePacketConn because... well... It's just not implemented yet:
https://github.com/golang/go/blob/964639cc338db650ccadeafb7424bc8ebb2c0f6c/src/net/file_windows.go
See also these issues:
https://github.com/golang/go/issues/9503
https://github.com/golang/go/issues/9661
*/

func DialBroadcastUDP(port int) net.PacketConn {
	s, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})

	f := os.NewFile(uintptr(s), "")
	conn, _ := net.FilePacketConn(f)
	f.Close()

	return conn
}
