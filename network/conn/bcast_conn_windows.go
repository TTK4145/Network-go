// +build windows

package conn

// Windows socket error codes can be found here
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms740668(v=vs.85).aspx

/*
Adventures in creating a broadcast socket for Go on Windows:

Alternative 1: The correct way that should work
To create a broadcast socket, you must first create a socket, then set the BROADCAST and REUSEADDR options, then call bind. 
However, the net.Dial/.Listen functions don't let you insert the calls to setsockopt between making the socket and binding 
it, because reasons.

Alternative 2: Syscalls
Instead of using the net package, we use syscall and just do it the proper way, and turn the file descriptor / handle into a 
"file connection". This works fine (as in "it works", not "it's fine" - because it is stupid) on posix, but does not work on 
Windows because reasons, where "reasons" are *it literally just says TODO in the standard library*:
https://github.com/golang/go/blob/ef0b09c526d78de23186522d50ff93ad657014c0/src/net/file_windows.go

Alternative 3: Golang shim of Windows API
The next option is to use sys/x/windows to interface with the Windows API directly:
https://godoc.org/golang.org/x/sys/windows
Which would work fine - and does work fine for most of the API - until again...  It's just not implemented yet:
https://github.com/golang/sys/blob/37707fdb30a5b38865cfb95e5aab41707daec7fd/windows/syscall_windows.go#L973-L980
Update 2022: This was added in December 2019: https://github.com/golang/sys/commit/6d18c012aee9febd81bbf9806760c8c4480e870d
But x/sys/windows is still an external package, which means a separate "go get" call and a flimsy dependency on GOROOT/GOPATH

Alternative 4: WSASockets from the Windows API
WSA Sockets are like normal sockets, but with more options and more parameters. I was not able to make them work though...
These are part of internal/syscall/windows, but go does not let users import internal packages

Alternative 5: Just write C code
This is what done up to 2022. All the socket code was written in C, and called from Go by using CGO, which therefore required 
a C compiler.

Alternative 6: Use ListenConfig, added in Go 1.11 (August 2018)
This is what is done below. A ListenConfig takes a callback that takes a callback that can operate on the file descriptor, 
which is then called by the ListenPacket function when setting up the socket. Absolutely not intuitive, and definitely not 
"the standard way", but it works.
This eliminates the need for a C compiler
*/



import (
    "context"
	"fmt"
	"net"
	"syscall"
)

func DialBroadcastUDP(port int) net.PacketConn {
    config := &net.ListenConfig{Control: 
        func (network, address string, conn syscall.RawConn) error {
            return conn.Control(func(descriptor uintptr) {
                syscall.SetsockoptInt(syscall.Handle(descriptor), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
                syscall.SetsockoptInt(syscall.Handle(descriptor), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
            })
        },
    }

	conn, err := config.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", port)) 
	if err != nil { fmt.Println("Error: net.ListenConfig.ListenPacket:", err) }

	return conn
}
