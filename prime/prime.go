package main

import (
    "math"
    "math/big"
    "bufio"
    "fmt"
    "net"
    "io"
    "encoding/json"
)

type JSONRequest struct {
    Method string `json:"method"`
    Number *float64 `json:"number"`
}

func IsPrime(value float64) bool {
	prime := false
	if math.Floor(value) == value {
		prime = big.NewInt(int64(value)).ProbablyPrime(0)
	}
	return prime
}

func prime(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn)
    for {
            data, err := reader.ReadString('\n')
            if err != nil {
                if err != io.EOF {
                    fmt.Println("read error:", err)
                }
                break;
            }

            var jsonReq JSONRequest
            err = json.Unmarshal([]byte(data), &jsonReq)
            if(err != nil || jsonReq.Method != "isPrime" || jsonReq.Number == nil) {
		conn.Write([]byte("{p\n"))
                continue
            }

	    isPrime := IsPrime(*jsonReq.Number)
	    jsonRes := struct {
		    Method string `json:"method"`
		    Prime bool `json:"prime"`
	    }{
		Method: "isPrime",
		Prime: isPrime,
	    }

	    resBytes, _ := json.Marshal(jsonRes)
	    conn.Write(append(resBytes, '\n'))
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

            defer conn.Close()
            go prime(conn)
        }
}
    

