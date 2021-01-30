package server

import (
	"fmt"
	"github.com/RinesThaix/homeTask/event"
)

type Handler struct {
	server *Server
}

func (h *Handler) Handle(rawEvent event.ClientEvent) (event.Event, error) {
	switch e := rawEvent.(type) {
	case *event.ClientInitialize:
		version, state := h.server.versioner.GetCurrentState()
		return &event.ServerInitializationResponse{Array: state, Version: version}, nil
	case *event.ClientAskForDiff:
		diff, err := h.server.versioner.GetOperationsSince(e.Version)
		if err != nil {
			return nil, err
		}
		return &event.ServerDiffResponse{Diff: diff}, nil
	case *event.ClientOperation:
		op := e.Operation.Copy() // because we don't really have any networking and (de)serialization, this exact field will be used on client for rollback actions
		rollback, diff, err := h.server.versioner.ProcessOperation(e.Version, op)
		return &event.ServerOperationResponse{Rollback: rollback, Diff: diff}, err
	default:
		return nil, fmt.Errorf("unknown event: %T", e)
	}
}
