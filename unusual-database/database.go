package main

import (
    "strings"
    "fmt"
    "net"
)

var store = make(map[string]string)

func handler(conn net.PacketConn, addr net.Addr, line []byte) {
    msg := strings.Trim(string(line), "\n")

    v := strings.SplitN(msg, "=", 2)
    key := v[0]
    if(len(v) > 1) { // Insert/Update values
        value := v[1]
        if(key != "version") {
            store[key] = value
        }
    } else { // Retrieve
        value, ok := store[key]

        if(ok) {
            fmt.Println("Found value for " + key + " : " + value)
            _, err := conn.WriteTo([]byte(key + "=" + value), addr)
            if(err != nil) {
                fmt.Println(err)
            }
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
    

