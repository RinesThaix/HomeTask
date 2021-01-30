package server

import (
	"context"
	"github.com/RinesThaix/homeTask/connection"
	"github.com/RinesThaix/homeTask/state"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type Server struct {
	Handler     *Handler
	versioner   *state.Versioner

	connectionID     int
	connections      map[int]*connection.ClientConnection
	connectionsMutex sync.Mutex
}

func NewServer(initialArraySize int) *Server {
	srv := &Server{}
	srv.Handler = &Handler{server: srv}
	srv.versioner = state.NewVersioner(state.NewState(srv.initArray(initialArraySize)), 1000)
	srv.connections = make(map[int]*connection.ClientConnection)
	srv.connectionsMutex = sync.Mutex{}
	return srv
}

func (s *Server) Initialize() {
	initBroadcaster(context.Background(), s, time.Millisecond * 500)
}

func (s *Server) OnClientConnected(conn *connection.ClientConnection) {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	s.connectionID++
	s.connections[s.connectionID] = conn
}

func (s *Server) OnClientDisconnected(conn *connection.ClientConnection) {
	// not storing client id anywhere except for the server itself, so just iterating over all values
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for id, c := range s.connections {
		if c == conn {
			delete(s.connections, id)
			return
		}
	}
}

func (s *Server) ProcessConnections(execution func(conn *connection.ClientConnection) error) error {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for _, conn := range s.connections {
		if err := execution(conn); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Array() []int32 {
	return s.versioner.State.Copy()
}

func (s *Server) initArray(size int) []int32 {
	array := make([]int32, size)
	for i := 0; i < size; i++ {
		array[i] = int32(rand.Uint32())
	}
	sort.Slice(array, func(i, j int) bool {
		return array[i] < array[j]
	})
	return array
}
