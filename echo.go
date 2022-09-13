package main

import (
        "bufio"
        "fmt"
        "net"
)

func connHandler(conn net.Conn) {
    for {
            netData, err := bufio.NewReader(conn).ReadString('\n')
            if err != nil {
                    fmt.Println(err)
                    return
            }
            fmt.Printf("%s -> %s", conn.RemoteAddr(), string(netData))
            conn.Write([]byte(netData))
            fmt.Printf("%s <- %s", conn.RemoteAddr(), string(netData))
    }
}

func main() {

        PORT := ":10000"
        l, err := net.Listen("tcp", PORT)
        if err != nil {
                fmt.Println(err)
                return
        }
        defer l.Close()

        for {
            conn, err := l.Accept()
            if err != nil {
                    fmt.Println(err)
                    continue
            }

            go connHandler(conn)
        }
}
    

