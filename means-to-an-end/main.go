package main

import (
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

    msgType, _ := reader.ReadByte()

    slice := []byte{}
    for i:=0; i < 4; i++ {
        r, _ := reader.ReadByte()
        slice = append(slice, r)
    }
    var num1 int32
    buf := bytes.NewBuffer(slice)
    binary.Read(buf, binary.BigEndian, &num1)

    slice2 := []byte{}
    for i:=0; i < 4; i++ {
        r, _ := reader.ReadByte()
        slice2 = append(slice2, r)
    }
    var num2 int32
    buf2 := bytes.NewBuffer(slice2)
    binary.Read(buf2, binary.BigEndian, &num2)


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
    var timestamps = make(map[int32]bool)

    reader := bufio.NewReader(conn)
    for {
        msg, err := readMessage(reader)
        if err != nil {
	    continue
        }

        if(msg.Type == 'I') {
            timestamp := msg.Number1 // Timestamp in seconds since 1970..
            price := msg.Number2 // Price in pennies

	    _, ok := timestamps[timestamp]
	    if ok {
		    conn.Write([]byte("{p\n"))
		    continue
	    } else {
		    timestamps[timestamp] = true
	    }
	    
            record := PriceRecord{
                Timestamp: timestamp,
                Price: price,
            }
            priceData = append(priceData, record)
        } else if(msg.Type == 'Q') {
            mintime := msg.Number1 // Timestamp in seconds since 1970..
            maxtime := msg.Number2 

            var sum int64 = 0
            var items int32 = 0
            for _, r := range priceData {
                if(r.Timestamp >= mintime && r.Timestamp <= maxtime) {
                    sum += int64(r.Price)
                    items++
                }
            }

	    var average int32 = 0
	    if (items > 0) {
		    average = int32(sum / int64(items))
		    //fmt.Println("Query res:", average)
	    }

            result := bytes.NewBuffer([]byte{})
            binary.Write(result, binary.BigEndian, average)
            conn.Write(result.Bytes())
        } else {
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
    

