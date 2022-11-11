package main

import (
    "strings"
    "fmt"
    "net"
)

var store = make(map[string]string)

func handler(conn net.PacketConn, addr net.Addr, line []byte) {
    msg := string(line)

    if(strings.Contains(msg, "=")) { // Insert/Update values
	if(msg[0] == '=') {
		store[""] = string(msg[1:])
	} else {
		key_val := strings.SplitN(msg, "=", 2)

		if(key_val[0] != "version") {
		    store[key_val[0]] = key_val[1]
		}
	}
    } else { // Retrieve
	res := msg + "="
	res += store[msg]

        _, err := conn.WriteTo([]byte(res), addr)
        if(err != nil) {
	    fmt.Println(err)
        }
    }
}


func main() {
        store["version"] = "version=Pluppen's Key-Value Store 1.0" // Init

        PORT := "10000"
        conn, err := net.ListenPacket("udp", "0.0.0.0:"+PORT)
        if err != nil {
		fmt.Println(err)
		return
        }

        defer conn.Close()
        for {
            buffer := make([]byte, 1000)
            n, addr, err := conn.ReadFrom(buffer)
            if(err != nil) {
                fmt.Println(err)
            }
            handler(conn, addr, buffer[:n])
        }
}
    

