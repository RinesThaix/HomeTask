package event

import "github.com/RinesThaix/homeTask/state"

type (
	Event interface{}

	ClientEvent interface {
		Event
	}

	ServerEvent interface {
		Event
	}
)

type (
	ClientInitialize struct {
		ClientEvent
	}

	ServerInitializationResponse struct {
		Event
		Version int
		Array   []int32
	}

	ClientAskForDiff struct {
		ClientEvent
		Version int
	}

	ServerDiffResponse struct {
		Event
		Diff []state.Operation
	}

	ClientOperation struct {
		ClientEvent
		Version   int
		Operation state.Operation
	}

	ServerOperationResponse struct {
		Event
		Rollback bool
		Diff     []state.Operation
	}
)

type (
	ServerDiff struct {
		ServerEvent
		Version int
		Diff    []state.Operation
	}
)
