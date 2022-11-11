package main

import (
    "strings"
    "bufio"
    "fmt"
    "net"
    "io"
    "sync"
    "regexp"  
)

var connCount int = 0;

type Session struct {
    Username string
    Connection net.Conn
}

func sendMessage(sessions *map[string]Session, username string, message string) {
    fmt.Println("--> " + message)
    for _, session := range *sessions {
        if(session.Username != username) {
            session.Connection.Write([]byte(message))
        }
    }
}

func handleSession(username string, reader *bufio.Reader, sessions *map[string]Session, conn net.Conn) {
    sendMessage(sessions, username, "* " + username + " has entered the room\n")

    for {
        data, err := reader.ReadString('\n')
        if err != nil {
            if err != io.EOF {
                fmt.Println("read error:", err)
            }
            break
        }
        sendMessage(sessions, username, "[" + username + "] " + data)
    }

    sendMessage(sessions, username, "* " + username + " has left the room\n")
    delete(*sessions, username)
}

func chat(sessions *map[string]Session, conn net.Conn, mutex *sync.Mutex) {
    defer conn.Close()

    var username string

    reader := bufio.NewReader(conn)

    conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
    data, err := reader.ReadString('\n')
    if err != nil {
        if err != io.EOF {
            fmt.Println("read error:", err)
        }
    }
    username = strings.Trim(data, "\n")

    var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString  
    if(!isStringAlphabetic(username) || len(username) == 0) {
        conn.Write([]byte("* Your username may only contain alphanumeric characters and must be of minimum length 1.\n"))
	fmt.Println("* Your username may only contain alphanumeric characters and must be of minimum length 1.\n")
        return
    }

    mutex.Lock()

    _, ok := (*sessions)[username]
    if(ok) {
	fmt.Println("--> * The username " + username + " is already taken.\n")
        conn.Write([]byte("* The username " + username + " is already taken.\n"))
        mutex.Unlock()
        return
    }

    usernames := []string{}
    for _, session := range *sessions {
        usernames = append(usernames, session.Username)
    }

    conn.Write([]byte("* The room contains: " + strings.Join(usernames, ", ") + "\n"))
    (*sessions)[username] = Session{Username: username, Connection: conn}

    mutex.Unlock()

    handleSession(username, reader, sessions, conn)
}


func main() {
        PORT := ":10000"
        l, err := net.Listen("tcp", PORT)
        if err != nil {
                fmt.Println(err)
                return
        }
        defer l.Close()

        var sessions = make(map[string]Session)
        var mutex sync.Mutex

        for {
            conn, err := l.Accept()
            if err != nil {
                fmt.Println(err)
                continue
            }

            defer conn.Close()
            go chat(&sessions, conn, &mutex)
        }
}
