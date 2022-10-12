package main

import (
    "io"
    "strings"
    "bufio"
    "fmt"
    "net"
)

var store = make(map[string]string)

func handler(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn)
    for {
        msg, err := reader.ReadBytes('\n')
        if(err != nil) {
            if(err != io.EOF) {
                fmt.Printf("%s\n", err)
            }
            break
        }
        msgStr := strings.Trim(string(msg), "\n")

        v := strings.SplitN(msgStr, "=", 2)
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
                _, err := conn.Write([]byte(key + "=" + value))
                if(err != nil) {
                    fmt.Println(err)
                }
            }
        }
    }
}


func main() {
        store["version"] = "version=Pluppen's Key-Value Store 1.0" // Init

        PORT := 10000
        addr := net.UDPAddr{
            Port: PORT,
            IP: net.ParseIP("127.0.0.1"),
        }
        conn, err := net.ListenUDP("udp", &addr)
        if err != nil {
                fmt.Println(err)
                return
        }

        defer conn.Close()
        for {
            handler(conn)
        }
}
    

