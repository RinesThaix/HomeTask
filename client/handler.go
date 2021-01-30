package client

import (
	"fmt"
	"github.com/RinesThaix/homeTask/event"
)

type Handler struct {
	client *Client
}

func (h *Handler) Handle(rawEvent event.ServerEvent) error {
	switch e := rawEvent.(type) {
	case *event.ServerDiff:
		h.client.mutex.Lock()
		defer h.client.mutex.Unlock()
		if h.client.awaitingResponse {
			return nil
		}
		clientVersion := h.client.version
		if clientVersion >= e.Version + len(e.Diff) {
			return nil
		}
		if e.Version > clientVersion {
			fmt.Printf("i'm too out of date, requesting more changes\n")
			errors := make(chan error)
			h.client.conn.SendWithCallback(&event.ClientAskForDiff{Version: clientVersion}, func(rawEvent event.Event, err error) {
				defer close(errors)
				if err != nil {
					errors <- err
					return
				}
				casted, ok := rawEvent.(*event.ServerDiffResponse)
				if !ok {
					errors <- fmt.Errorf("received unexpected response for diff request: %T", rawEvent)
					return
				}
				e.Diff = casted.Diff
				e.Version = clientVersion
			})
			for err := range errors {
				if err != nil {
					return err
				}
			}
		}
		if err := h.client.state.PerformMany(e.Diff, clientVersion - e.Version); err != nil {
			return fmt.Errorf("could not apply server diff: %w", err)
		}
		h.client.version = e.Version + len(e.Diff)
		return nil
	default:
		return fmt.Errorf("unknown event: %T", e)
	}
}
