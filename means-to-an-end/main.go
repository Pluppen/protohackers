package main

import (
    "math"
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

type PriceRecord struct {
    Timestamp int32
    Price int32
}

func readMessage(reader *bufio.Reader) (Message, error) {
    res := Message{}

    msgType, err := reader.ReadByte()
    if(err != nil) {
        return res, err
    }

    var slice = []byte{}
    for i:=0; i < 4; i++ {
        r, err := reader.ReadByte()
        if(err != nil) {
            return res, err
        }

        slice = append(slice, r)
    }
    var num1 int32
    buf := bytes.NewBuffer(slice)
    binary.Read(buf, binary.BigEndian, &num1)

    slice = []byte{}
    for i:=0; i < 4; i++ {
        r, err := reader.ReadByte()
        if(err != nil) {
            return res, err
        }
        slice = append(slice, r)
    }
    var num2 int32
    buf = bytes.NewBuffer(slice)
    binary.Read(buf, binary.BigEndian, &num2)


    res = Message{
        Type: msgType,
        Number1: num1,
        Number2: num2,
    }
    return res, nil
}

func handler(conn net.Conn) {
    defer conn.Close()

    var priceData []PriceRecord

    reader := bufio.NewReader(conn)
    for {
        msg, err := readMessage(reader)
        if err != nil {
            fmt.Println("read error:", err)
            break;
        }

        if(msg.Type == 'I') {
            timestamp := msg.Number1 // Timestamp in seconds since 1970..
            price := msg.Number2 // Price in pennies
            record := PriceRecord{
                Timestamp: timestamp,
                Price: price,
            }
            fmt.Println("Inserted ", price, " at timestamp ", timestamp)
            priceData = append(priceData, record)
        } else if(msg.Type == 'Q') {
            mintime := msg.Number1 // Timestamp in seconds since 1970..
            maxtime := msg.Number2 

            var sum int32 = 0
            var items int32 = 0
            for _, r := range priceData {
                if(r.Timestamp >= mintime && r.Timestamp <= maxtime) {
                    sum += r.Price
                    items++
                }
            }
            fmt.Println("Query res:", sum/items)

            result := bytes.NewBuffer([]byte{})
            binary.Write(result, binary.BigEndian, math.Round(float64(sum / items)))
            conn.Write(result.Bytes())
        } else {
            fmt.Println("Something went wrong: ",msg)
            conn.Write([]byte("{p\n"))
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
    

