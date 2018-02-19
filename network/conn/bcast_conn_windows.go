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
https://github.com/golang/go/blob/dfb0e4f6c744eb9bf629658bf7da313b2d1518e1/src/net/file_windows.go

Alternative 3: Golang shim of Windows API
The next option is to use sys/x/windows to interface with the Windows API directly:
https://godoc.org/golang.org/x/sys/windows
Which would work fine - and does work fine for most of the API - until again...  It's just not implemented yet:
https://github.com/golang/sys/blob/37707fdb30a5b38865cfb95e5aab41707daec7fd/windows/syscall_windows.go#L973-L980

Alternative 4: WSASockets from the Windows API
WSA Sockets are like normal sockets, but with more options and more parameters. I was not able to make them work though...

Alternative 5: Just write C code
This is what is done below. All the socket code is written in C, and called from Go by using CGO, which therefore requires 
a C compiler.
*/



/*
#include<stdio.h>
#include<winsock2.h>


SOCKET cBcastSocket(u_short port){
    SOCKET s;

    if((s = socket(AF_INET , SOCK_DGRAM , 0 )) == INVALID_SOCKET){
        printf("Could not create socket: %d\n" , WSAGetLastError());
        exit(EXIT_FAILURE);
    }

    int opt = 1;
    int optlen = sizeof(opt);
    if(setsockopt(s, SOL_SOCKET, SO_BROADCAST, (char*)&opt, optlen) == SOCKET_ERROR){
        printf("SO_BROADCAST failed with error code : %d\n" , WSAGetLastError());
        exit(EXIT_FAILURE);
    }
    if(setsockopt(s, SOL_SOCKET, SO_REUSEADDR, (char*)&opt, optlen) == SOCKET_ERROR){
        printf("SO_REUSEADDR failed with error code : %d\n" , WSAGetLastError());
        exit(EXIT_FAILURE);
    }

    struct sockaddr_in a;
    a.sin_family = AF_INET;
    a.sin_addr.s_addr = INADDR_ANY;
    a.sin_port = htons(port);

    if(bind(s, (struct sockaddr*)&a, sizeof(a)) == SOCKET_ERROR){
        printf("Bind failed with error code : %d\n" , WSAGetLastError());
        exit(EXIT_FAILURE);
    }

    return s;
}

int cSendTo(SOCKET s, char* ip, u_short port, char* buf, int buflen){
    struct sockaddr_in a;
    a.sin_family = AF_INET;
    a.sin_addr.s_addr = inet_addr(ip);
    a.sin_port = htons(port);
    return sendto(s, buf, buflen, 0, (struct sockaddr*)&a, sizeof(struct sockaddr));
}

int cRecvFrom(SOCKET s, char* buf, int buflen, char* addr){
    struct sockaddr_in* a = (struct sockaddr_in*)addr;
    int a_len = sizeof(struct sockaddr);
    return recvfrom(s, buf, buflen, 0, (struct sockaddr *)a, &a_len);
}

int cClose(SOCKET s){
    return closesocket(s);
}

int cLocalAddr(SOCKET s, char* addr){
    struct sockaddr_in* a = (struct sockaddr_in*)addr;
    int a_len = sizeof(struct sockaddr);
    return getsockname(s, (struct sockaddr *)a, &a_len);
}

int cSetReadDeadline(SOCKET s, int timeout_ms){
    return setsockopt(s, SOL_SOCKET, SO_RCVTIMEO, (char*)&timeout_ms, sizeof(int));
}

int cSetWriteDeadline(SOCKET s, int timeout_ms){
    return setsockopt(s, SOL_SOCKET, SO_SNDTIMEO, (char*)&timeout_ms, sizeof(int));
}

 #cgo LDFLAGS: -lws2_32
*/
import "C"
import (
	"errors"
	"fmt"
	"net"
	"time"
	"unsafe"
)

type WindowsBroadcastConn struct {
	Sock C.SOCKET
}

func (f WindowsBroadcastConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	var addrbuf [16]byte
	r := int(C.cRecvFrom(f.Sock, (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)), (*C.char)(unsafe.Pointer(&addrbuf[0]))))
	if r == C.SOCKET_ERROR {
		return 0, nil, errors.New(fmt.Sprintf("recvfrom() failed with error code %d", C.WSAGetLastError()))
	} else {
		addr := net.UDPAddr{IP: addrbuf[4:8], Port: int(C.ntohs(C.u_short(addrbuf[2]<<8) | C.u_short(addrbuf[3])))}
		return r, &addr, nil
	}
}

func (f WindowsBroadcastConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	r := int(C.cSendTo(f.Sock, C.CString(addr.(*net.UDPAddr).IP.String()), C.u_short(addr.(*net.UDPAddr).Port), (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b))))
	if r == C.SOCKET_ERROR {
		return 0, errors.New(fmt.Sprintf("sendto() failed with error code %d", C.WSAGetLastError()))
	} else {
		return r, nil
	}
}

func (f WindowsBroadcastConn) Close() error {
	r := C.cClose(f.Sock)
	if r == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("closesocket() failed with error code %d", C.WSAGetLastError()))
	}
}

func (f WindowsBroadcastConn) LocalAddr() net.Addr {
	var addrbuf [16]byte
	r := int(C.cLocalAddr(f.Sock, (*C.char)(unsafe.Pointer(&addrbuf[0]))))
	if r == C.SOCKET_ERROR {
		return nil
	} else {
		addr := net.UDPAddr{IP: addrbuf[4:8], Port: int(C.ntohs(C.u_short(addrbuf[2]<<8) | C.u_short(addrbuf[3])))}
		return &addr
	}
}

func (f WindowsBroadcastConn) SetDeadline(t time.Time) error {
	e := f.SetReadDeadline(t)
	if e != nil {
		return e
	}
	e = f.SetWriteDeadline(t)
	return e
}

func (f WindowsBroadcastConn) SetReadDeadline(t time.Time) error {
	timeout_ms := int64(t.Sub(time.Now())) / 1000000
	r := -1
	if timeout_ms > 0 {
		r = int(C.cSetReadDeadline(f.Sock, C.int(timeout_ms)))
	}
	if r == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("setsockopt() failed with error code %d", C.WSAGetLastError()))
	}
}

func (f WindowsBroadcastConn) SetWriteDeadline(t time.Time) error {
	timeout_ms := int64(t.Sub(time.Now())) / 1000000
	r := -1
	if timeout_ms > 0 {
		r = int(C.cSetWriteDeadline(f.Sock, C.int(timeout_ms)))
	}
	if r == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("setsockopt() failed with error code %d", C.WSAGetLastError()))
	}
}

func DialBroadcastUDP(port int) net.PacketConn {
	return WindowsBroadcastConn{C.cBcastSocket(C.u_short(port))}
}
