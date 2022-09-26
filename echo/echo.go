package main

import (
    "fmt"
    "net"
    "io"
)

func echo(conn net.Conn) {
    defer conn.Close()

    buf := make([]byte, 0, 4096)
    tmp := make([]byte, 256)

    for {
            data, err := conn.Read(tmp)
            if err != nil {
                if err != io.EOF {
                    fmt.Println("read error:", err)
                }
                break;
            }
            buf = append(buf, tmp[:data]...)
    }
    fmt.Printf("%s -> %s", conn.RemoteAddr(), string(buf))

    fmt.Printf("%s <- %s", conn.RemoteAddr(), string(buf))
    conn.Write([]byte(buf))
}

func connHandler(conn net.Conn) {
    defer conn.Close()
    echo(conn)
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

            defer conn.Close()
            go echo(conn)
        }
}
    

