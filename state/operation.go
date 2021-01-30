package state

import (
	"fmt"
	"strings"
)

type (
	Operation interface {
		fmt.Stringer
		Copy() Operation
	}

	OpInsert struct {
		Position int
		Value    int32
	}

	OpUpdate struct {
		Position      int
		Value         int32
		PreviousValue int32
	}

	OpDelete struct {
		Position      int
		PreviousValue int32
	}

	OpBatch struct {
		Operations []Operation
	}
)

func (op *OpInsert) Copy() Operation {
	return &OpInsert{Position: op.Position, Value: op.Value}
}

func (op *OpInsert) String() string {
	return fmt.Sprintf("insert{pos=%d,value=%d}", op.Position, op.Value)
}

func (op *OpUpdate) Copy() Operation {
	return &OpUpdate{Position: op.Position, Value: op.Value, PreviousValue: op.PreviousValue}
}

func (op *OpUpdate) String() string {
	return fmt.Sprintf("update{pos=%d,value=%d}", op.Position, op.Value)
}

func (op *OpDelete) Copy() Operation {
	return &OpDelete{Position: op.Position, PreviousValue: op.PreviousValue}
}

func (op *OpDelete) String() string {
	return fmt.Sprintf("delete{pos=%d}", op.Position)
}

func (op *OpBatch) Copy() Operation {
	operations := make([]Operation, len(op.Operations))
	copy(operations, op.Operations)
	return &OpBatch{Operations: operations}
}

func (op *OpBatch) String() string {
	var children []string
	for _, o := range op.Operations {
		children = append(children, o.String())
	}
	return fmt.Sprintf("batch{%s}", strings.Join(children, ","))
}
