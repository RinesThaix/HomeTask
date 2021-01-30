package connection

import (
	"github.com/RinesThaix/homeTask/event"
)

type (
	ClientConnection struct {
		SendFunc func(event event.ServerEvent)
	}

	ServerConnection struct {
		SendFunc func(event event.ClientEvent) (event.Event, error)
	}
)

func (c *ClientConnection) Send(event event.ServerEvent) {
	go c.SendFunc(event)
}

func (c *ServerConnection) Send(event event.ClientEvent) {
	c.SendWithCallback(event, nil)
}

func (c *ServerConnection) SendWithCallback(event event.ClientEvent, callback func(event event.Event, err error)) {
	go func() {
		response, err := c.SendFunc(event)
		if callback == nil {
			if err != nil {
				panic(err)
			}
		} else {
			callback(response, err)
		}
	}()
}
