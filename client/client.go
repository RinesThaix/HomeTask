package client

import (
	"fmt"
	"github.com/RinesThaix/homeTask/connection"
	"github.com/RinesThaix/homeTask/event"
	"github.com/RinesThaix/homeTask/server"
	"github.com/RinesThaix/homeTask/state"
	"sync"
)

type Client struct {
	Handler *Handler
	server  *server.Server
	state   *state.State
	conn    *connection.ServerConnection

	offlineOperations []state.Operation

	clientConn       *connection.ClientConnection
	version          int
	awaitingResponse bool
	mutex            sync.Mutex
}

func NewClient(server *server.Server) *Client {
	c := &Client{}
	c.server = server
	c.Handler = &Handler{client: c}
	c.state = state.NewState(make([]int32, 0))
	c.conn = &connection.ServerConnection{SendFunc: server.Handler.Handle}
	return c
}

func (c *Client) Initialize() error {
	return c.initialize(false)
}

func (c *Client) initialize(locked bool) error {
	errors := make(chan error)
	c.conn.SendWithCallback(&event.ClientInitialize{}, func(rawEvent event.Event, err error) {
		defer close(errors)
		if err != nil {
			errors <- fmt.Errorf("could not initialize client: %w", err)
			return
		}
		casted, ok := rawEvent.(*event.ServerInitializationResponse)
		if !ok {
			errors <- fmt.Errorf("received unexpected response to client initialization: %T", rawEvent)
			return
		}
		c.state.Set(casted.Array)
		c.mutex.Lock()
		c.version = casted.Version
		c.mutex.Unlock()
	})
	for err := range errors {
		if err != nil {
			return err
		}
	}
	if !locked {
		c.mutex.Lock()
		defer c.mutex.Unlock()
	}
	c.clientConn = &connection.ClientConnection{SendFunc: func(event event.ServerEvent) {
		if err := c.Handler.Handle(event); err != nil {
			panic(err)
		}
	}}
	c.server.OnClientConnected(c.clientConn) // imitating handshaking
	c.sendOfflineChanges()
	return nil
}

func (c *Client) Disconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.disconnect()
}

func (c *Client) disconnect() {
	if c.clientConn == nil {
		return
	}
	c.server.OnClientDisconnected(c.clientConn)
	c.version = 0
	c.awaitingResponse = false
	c.clientConn = nil
}

func (c *Client) reinitialize() error {
	c.disconnect()
	return c.initialize(true)
}

func (c *Client) Insert(pos int, value int32) error {
	return c.modify(&state.OpInsert{Position: pos, Value: value})
}

func (c *Client) Update(pos int, value int32) error {
	previousValue, err := c.state.Get(pos)
	if err != nil {
		return err
	}
	return c.modify(&state.OpUpdate{Position: pos, Value: value, PreviousValue: previousValue})
}

func (c *Client) Delete(pos int) error {
	previousValue, err := c.state.Get(pos)
	if err != nil {
		return err
	}
	return c.modify(&state.OpDelete{Position: pos, PreviousValue: previousValue})
}

func (c *Client) Batch(operations []state.Operation) error {
	return c.modify(&state.OpBatch{Operations: operations})
}

func (c *Client) Get(pos int) (int32, error) {
	return c.state.Get(pos)
}

func (c *Client) Array() []int32 {
	return c.state.Copy()
}

func (c *Client) Size() int {
	return c.state.Size()
}

func (c *Client) modify(op state.Operation) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.clientConn == nil {
		// offline mode
		if err := c.state.Perform(op); err != nil {
			return err
		}
		c.offlineOperations = append(c.offlineOperations, op)
		return nil
	}
	conn := c.clientConn
	if c.awaitingResponse {
		// actually we can delay operation there and not just die
		fmt.Printf("awaiting response from the server for previous operation\n")
		return nil
	}
	if err := c.state.Perform(op); err != nil {
		return err
	}
	version := c.version
	c.awaitingResponse = true
	c.conn.SendWithCallback(&event.ClientOperation{Version: c.version, Operation: op}, func(rawEvent event.Event, err error) {
		c.mutex.Lock()
		defer func() {
			c.awaitingResponse = false
			c.mutex.Unlock()
		}()
		if conn != c.clientConn {
			// reinitialized
			return
		}
		if version != c.version {
			// received update from the broadcaster
			return
		}
		if err != nil {
			fmt.Printf("received error response to client operation %v: %v\n", op, err)
		}
		casted, ok := rawEvent.(*event.ServerOperationResponse)
		if !ok {
			panic(fmt.Errorf("received unexpected response to client operation: %T", rawEvent))
		}

		if casted.Rollback {
			err = c.state.RollbackAndPerformMany(op, casted.Diff)
			if casted.Diff != nil {
				c.version += len(casted.Diff)
			}
		} else {
			err = c.state.PerformMany(casted.Diff, 0)
			if casted.Diff != nil {
				c.version += len(casted.Diff)
			}
			c.version++
		}

		if err != nil {
			panic(fmt.Errorf("could not process server's response to operation: %w", err))
		}
	})
	return nil
}

func (c *Client) sendOfflineChanges() error {
	if len(c.offlineOperations) == 0 {
		return nil
	}
	if err := c.modify(&state.OpBatch{Operations: c.offlineOperations}); err != nil {
		return err
	}
	c.offlineOperations = make([]state.Operation, 0)
	return nil
}
