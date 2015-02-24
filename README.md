Network-go
==========

Network packages for TTK4145 Real-time Programming term project. The initial versions of these are ment as a modules the students should collaborate on to further develop them 

Package Status:
------
####/udp/udp.go:
| Test	            | Status | File    |  	Comments		|
|-------------------|:------:|---------|--------------------------------|
| Broadcast loopback| Passed | example/test-udp-bloopbck.go 	| Prelimenary test 	|
	
| Issue | Status | Comments  |
|-------|--------|-----------|


####/tcp/tcp.go:
| Test	| Status | 	File |  	Comments					|
|-------------------|:------:|--------------------------------|-------------------------------|
| 2 way connection | Passed | example/test-tcp.go 	| Prelimenary test 						 	|
| Issue | Status | Comments  |
|-------|--------|-----------|
| Killing one side causes crash | Not resolved  | When one side of the connection is killed the other side crashes, should add checks for alive connections and close & delete dead connections | 
