package main

import (
    "math"
        "bufio"
        "fmt"
        "net"
        "io"
        "encoding/json"
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

type JSONRequest struct {
    Method string `json:"method"`
    Number int `json:"number"`
}

type JSONResponse struct {
    Method string `json:"method"`
    Prime bool `json:"prime"`
}

func IsPrime(value int) bool {
    for i := 2; i <= int(math.Floor(float64(value)/2)); i++ {
        if value%i == 0 {
            return false
        }
    }
    return value > 1
}

func prime(conn net.Conn) {
    defer conn.Close()

    for {
            data, err := bufio.NewReader(conn).ReadBytes('\n')
            if err != nil {
                if err != io.EOF {
                    fmt.Println("read error:", err)
                }
                break;
            }

            var jsonRes JSONResponse
            jsonRes.Method = "malformed"
            jsonRes.Prime = false

            var jsonReq JSONRequest
            err = json.Unmarshal(data, &jsonReq)
            if err != nil {
                fmt.Println(err)
            }
            fmt.Printf("%s <- %s", conn.RemoteAddr(), string(data))

            if(jsonReq.Method == "isPrime") {
                isPrime := IsPrime(jsonReq.Number)
                jsonRes.Method = jsonReq.Method
                jsonRes.Prime = isPrime
                json.NewEncoder(conn).Encode(jsonRes)
                jsonString, _ := json.Marshal(jsonRes)
                fmt.Printf("%s -> %s", conn.RemoteAddr(), string(jsonString))
                continue
            }

            jsonString, err := json.Marshal(jsonRes)
            fmt.Printf("%s -> %s", conn.RemoteAddr(), string(jsonString))
            json.NewEncoder(conn).Encode(jsonRes)
    }
}

func connHandler(conn net.Conn) {
    defer conn.Close()
    prime(conn)
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
    

