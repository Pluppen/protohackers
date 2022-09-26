package main

import (
    "errors"
    "io"
    "bytes"
    "bufio"
    "fmt"
    "net"
    "encoding/binary"
)

type Message struct {
    Type byte
    Number1 int32
    Number2 int32
}

func readMessage(reader *bufio.Reader) (Message, error) {
    res := Message{}

    message := make([]byte, 9)
    n, err := io.ReadFull(reader, message)
    if(err != nil) {
        return res, err
    }
    
    if(n != 9) {
        return res, errors.New("Could not read 9 bytes")
    }

    num1 := int32(binary.BigEndian.Uint32(message[1:5]))
    num2 := int32(binary.BigEndian.Uint32(message[5:9]))

    res = Message{
        Type: message[0],
        Number1: num1,
        Number2: num2,
    }
    return res, nil
}

func handler(conn net.Conn) {
    defer conn.Close()

    var priceRecords = make(map[int32]int32)

    reader := bufio.NewReader(conn)
    for {
        msg, err := readMessage(reader)
        if(err != nil) {
            if(err != io.EOF) {
                fmt.Printf("%s\n", err)
            }
            break
        }

        if(msg.Type == 'I') {
            timestamp := msg.Number1
            price := msg.Number2

            priceRecords[timestamp] = price
        } else if(msg.Type == 'Q') {
            mintime := msg.Number1
            maxtime := msg.Number2 

            var sum int64 = 0
            items := 0
            for timestamp, price := range priceRecords {
                if(timestamp >= mintime && timestamp <= maxtime) {
                    sum += int64(price)
                    items++
                }
            }

            var average int32 = 0
            if (items > 0) {
                average = int32(sum / int64(items))
            }

            result := bytes.NewBuffer([]byte{})
            binary.Write(result, binary.BigEndian, average)
            conn.Write(result.Bytes())
        } else {
            break
        }
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
            go handler(conn)
        }
}
    

