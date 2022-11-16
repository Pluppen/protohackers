package main

import (
    "bufio"
    "fmt"
    "net"
    "io"
    "strings"
    "regexp"  
)

const tonysAddress string = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
const serverEndpoint string = "chat.protohackers.com:16963"

type ProxyConnection struct {
    client net.Conn
    clientData chan string
    server net.Conn
    serverData chan string
}

func (proxy *ProxyConnection) Close() {
    proxy.server.Close();
    proxy.client.Close();
}

func (proxy *ProxyConnection) receiveFromServer() {
    reader := bufio.NewReader(proxy.server)
    for {
        message, err := reader.ReadString('\n')
        if(err != nil) {
            if(err != io.EOF) {
                fmt.Printf("%s\n", err)
            }
            break
        }
        proxy.serverData <- strings.TrimSpace(message)
    }
    close(proxy.serverData)
}

func (proxy *ProxyConnection) receiveFromClient() {
    reader := bufio.NewReader(proxy.client)
    for {
        message, err := reader.ReadString('\n')
        if(err != nil) {
            if(err != io.EOF) {
                fmt.Printf("%s\n", err)
            }
            break
        }
        proxy.clientData <- strings.TrimSpace(message)
    }
    close(proxy.clientData)
}

func createProxyConnection(client net.Conn) *ProxyConnection {
    upstream, err := net.Dial("tcp", serverEndpoint)
    if(err != nil) {
        fmt.Println("Could not connect to server endpoint.")
        panic(err)
    }

    proxy := ProxyConnection{
        client: client,
        clientData: make(chan string),
        server: upstream,
        serverData: make(chan string),
    }

    go proxy.receiveFromServer()
    go proxy.receiveFromClient()

    return &proxy
}

func (proxy *ProxyConnection) handleConnections() {
    for {
        select {
        case clientData := <-proxy.clientData:
            if clientData == "" {
                goto disconnect
            }
            fmt.Printf("client: %s\n", clientData)
            proxy.server.Write([]byte(rewrite(clientData) + "\n"))
        case serverData := <-proxy.serverData:
            if serverData == "" {
                goto disconnect
            }
            fmt.Printf("server: %s\n", serverData)
            proxy.client.Write([]byte(rewrite(serverData) + "\n"))
        }
    }
disconnect:
    proxy.Close()
}

func rewrite(message string) string {
    bogusRx := regexp.MustCompile("^7[0-9a-zA-Z]{25,34}$")

    s := strings.Split(message, " ")
    for i, word := range s {
        if bogusRx.MatchString(word) {
            s[i] = tonysAddress
        }
    }

    return strings.Join(s, " ")
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

            proxy :=  createProxyConnection(conn)
            go proxy.handleConnections()
        }
}
