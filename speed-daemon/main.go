package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// Constants for message types
const (
	MsgError         = 0x10
	MsgPlate         = 0x20
	MsgTicket        = 0x21
	MsgWantHeartbeat = 0x40
	MsgHeartbeat     = 0x41
	MsgIAmCamera     = 0x80
	MsgIAmDispatcher = 0x81
)

// Constants for time
const (
	SecondsPerDay = 86400
)

// Observation represents a single camera observation
type Observation struct {
	Plate     string
	Mile      uint16
	Timestamp uint32
	Road      uint16
	Limit     uint16
}

// Ticket represents a speed violation ticket
type Ticket struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	Speed      uint16
}

// Camera represents a speed camera client
type Camera struct {
	Road  uint16
	Mile  uint16
	Limit uint16
	Conn  net.Conn
}

// Dispatcher represents a ticket dispatcher client
type Dispatcher struct {
	Roads []uint16
	Conn  net.Conn
}

// Server represents the main server structure
type Server struct {
	sync.RWMutex
	cameras       map[string]*Camera
	dispatchers   map[uint16][]*Dispatcher
	observations  map[uint16]map[string][]Observation // road -> plate -> []Observation
	issuedTickets map[string]map[uint32]bool          // plate -> day -> issued
}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{
		cameras:       make(map[string]*Camera),
		dispatchers:   make(map[uint16][]*Dispatcher),
		observations:  make(map[uint16]map[string][]Observation),
		issuedTickets: make(map[string]map[uint32]bool),
	}
}

// Binary protocol reading helpers
func readU8(r io.Reader) (uint8, error) {
	var n uint8
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func readU16(r io.Reader) (uint16, error) {
	var n uint16
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func readU32(r io.Reader) (uint32, error) {
	var n uint32
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func readString(r io.Reader) (string, error) {
	length, err := readU8(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// Write helpers
func writeError(conn net.Conn, msg string) error {
	if err := binary.Write(conn, binary.BigEndian, uint8(MsgError)); err != nil {
		return err
	}
	if err := binary.Write(conn, binary.BigEndian, uint8(len(msg))); err != nil {
		return err
	}
	if _, err := conn.Write([]byte(msg)); err != nil {
		return err
	}
	return nil
}

func writeTicket(conn net.Conn, t *Ticket) error {
	if err := binary.Write(conn, binary.BigEndian, uint8(MsgTicket)); err != nil {
		return err
	}

	// Write plate length and plate
	if err := binary.Write(conn, binary.BigEndian, uint8(len(t.Plate))); err != nil {
		return err
	}
	if _, err := conn.Write([]byte(t.Plate)); err != nil {
		return err
	}

	// Write other fields
	fields := []interface{}{
		t.Road,
		t.Mile1,
		t.Timestamp1,
		t.Mile2,
		t.Timestamp2,
		t.Speed,
	}

	for _, field := range fields {
		if err := binary.Write(conn, binary.BigEndian, field); err != nil {
			return err
		}
	}

	return nil
}

// handleHeartbeat starts sending heartbeat messages at the specified interval
func (s *Server) handleHeartbeat(conn net.Conn, interval uint32) {
	if interval == 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * 100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if err := binary.Write(conn, binary.BigEndian, uint8(MsgHeartbeat)); err != nil {
			return
		}
	}
}

// calculateSpeed calculates the average speed between two observations
func calculateSpeed(mile1, mile2 uint16, timestamp1, timestamp2 uint32) uint16 {
	distance := float64(mile2 - mile1)
	timeSeconds := float64(timestamp2 - timestamp1)
	speedMPH := (distance / timeSeconds) * 3600
	return uint16(speedMPH * 100)
}

// checkViolation checks if a new observation creates a speed violation
func (s *Server) checkViolation(obs Observation) {
	s.Lock()
	defer s.Unlock()

	roadObs, exists := s.observations[obs.Road]
	if !exists {
		roadObs = make(map[string][]Observation)
		s.observations[obs.Road] = roadObs
	}

	plateObs := roadObs[obs.Plate]
	plateObs = append(plateObs, obs)
	roadObs[obs.Plate] = plateObs

	// Check for violations
	for i := 0; i < len(plateObs)-1; i++ {
		for j := i + 1; j < len(plateObs); j++ {
			obs1 := plateObs[i]
			obs2 := plateObs[j]

			if obs1.Timestamp >= obs2.Timestamp {
				continue
			}

			speed := calculateSpeed(obs1.Mile, obs2.Mile, obs1.Timestamp, obs2.Timestamp)
			if float64(speed)/100.0 >= float64(obs.Limit)+0.5 {
				// Check if ticket already issued for this day
				day := obs1.Timestamp / SecondsPerDay
				if s.issuedTickets[obs.Plate] == nil {
					s.issuedTickets[obs.Plate] = make(map[uint32]bool)
				}
				if !s.issuedTickets[obs.Plate][day] {
					ticket := &Ticket{
						Plate:      obs.Plate,
						Road:       obs.Road,
						Mile1:      obs1.Mile,
						Timestamp1: obs1.Timestamp,
						Mile2:      obs2.Mile,
						Timestamp2: obs2.Timestamp,
						Speed:      speed,
					}
					s.issueTicket(ticket)
					s.issuedTickets[obs.Plate][day] = true
				}
			}
		}
	}
}

// issueTicket sends a ticket to an appropriate dispatcher
func (s *Server) issueTicket(ticket *Ticket) {
	dispatchers := s.dispatchers[ticket.Road]
	if len(dispatchers) == 0 {
		return
	}

	// Send to first available dispatcher
	if err := writeTicket(dispatchers[0].Conn, ticket); err != nil {
		// Handle error, maybe try another dispatcher
		log.Printf("Failed to send ticket: %v", err)
	}
}

// handleConnection handles a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		msgType, err := readU8(conn)
		if err != nil {
			return
		}

		switch msgType {
		case MsgIAmCamera:
			road, err := readU16(conn)
			if err != nil {
				return
			}
			mile, err := readU16(conn)
			if err != nil {
				return
			}
			limit, err := readU16(conn)
			if err != nil {
				return
			}

			camera := &Camera{
				Road:  road,
				Mile:  mile,
				Limit: limit,
				Conn:  conn,
			}

			s.Lock()
			s.cameras[conn.RemoteAddr().String()] = camera
			s.Unlock()

		case MsgIAmDispatcher:
			numRoads, err := readU8(conn)
			if err != nil {
				return
			}

			roads := make([]uint16, numRoads)
			for i := 0; i < int(numRoads); i++ {
				road, err := readU16(conn)
				if err != nil {
					return
				}
				roads[i] = road
			}

			dispatcher := &Dispatcher{
				Roads: roads,
				Conn:  conn,
			}

			s.Lock()
			for _, road := range roads {
				s.dispatchers[road] = append(s.dispatchers[road], dispatcher)
			}
			s.Unlock()

		case MsgPlate:
			plate, err := readString(conn)
			if err != nil {
				return
			}
			timestamp, err := readU32(conn)
			if err != nil {
				return
			}

			s.RLock()
			camera, exists := s.cameras[conn.RemoteAddr().String()]
			s.RUnlock()

			if !exists {
				writeError(conn, "not a camera")
				return
			}

			obs := Observation{
				Plate:     plate,
				Mile:      camera.Mile,
				Timestamp: timestamp,
				Road:      camera.Road,
				Limit:     camera.Limit,
			}
			s.checkViolation(obs)

		case MsgWantHeartbeat:
			interval, err := readU32(conn)
			if err != nil {
				return
			}
			go s.handleHeartbeat(conn, interval)

		default:
			writeError(conn, "illegal msg")
			return
		}
	}
}

func main() {
	server := NewServer()

	listener, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go server.handleConnection(conn)
	}
}
