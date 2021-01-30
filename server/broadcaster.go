package server

import (
	"context"
	"fmt"
	"github.com/RinesThaix/homeTask/connection"
	"github.com/RinesThaix/homeTask/event"
	"time"
)

type broadcaster struct {
	server        *Server
	latestVersion int
}

func initBroadcaster(ctx context.Context, server *Server, interval time.Duration) {
	b := &broadcaster{server: server, latestVersion: 0}
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := b.broadcast(); err != nil {
					panic(fmt.Errorf("could not broadcast server diff: %w", err))
				}
			}
		}
	}()
}

func (b *broadcaster) broadcast() error {
	version := b.latestVersion
	operations, err := b.server.versioner.GetOperationsSince(version)
	if err != nil {
		return err
	}
	if len(operations) == 0 {
		return nil
	}
	b.latestVersion = version + len(operations)
	return b.server.ProcessConnections(func(conn *connection.ClientConnection) error {
		conn.Send(&event.ServerDiff{Version: version, Diff: operations})
		return nil
	})
}
