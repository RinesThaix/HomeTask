package state

import (
	"fmt"
	"github.com/RinesThaix/homeTask/util"
	"sync"
)

type State struct {
	LastOp Operation
	array  *util.BlockedArray
	mutex  sync.RWMutex
}

func NewState(initialArray []int32) *State {
	return &State{array: util.NewBlockedArray(initialArray, 10), mutex: sync.RWMutex{}}
}

func (s *State) Perform(operation Operation) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.perform(operation)
}

func (s *State) PerformMany(operations []Operation, offset int) error {
	if operations == nil {
		return nil
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for i := offset; i < len(operations); i++ {
		if err := s.perform(operations[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) RollbackAndPerformMany(rollback Operation, operations []Operation) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.rollback(rollback); err != nil {
		return err
	}
	if operations != nil {
		for _, op := range operations {
			if err := s.perform(op); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *State) perform(operation Operation) error {
	switch op := operation.(type) {
	case *OpInsert:
		if err := s.insert(op.Position, op.Value); err != nil {
			return err
		}
		s.LastOp = operation
		return nil
	case *OpUpdate:
		if err := s.update(op.Position, op.Value); err != nil {
			return err
		}
		s.LastOp = operation
		return nil
	case *OpDelete:
		if err := s.delete(op.Position); err != nil {
			return err
		}
		s.LastOp = operation
		return nil
	case *OpBatch:
		for _, el := range op.Operations {
			if err := s.perform(el); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown operation: %T", op)
	}
}

func (s *State) rollback(operation Operation) error {
	switch op := operation.(type) {
	case *OpInsert:
		return s.delete(op.Position)
	case *OpUpdate:
		return s.update(op.Position, op.PreviousValue)
	case *OpDelete:
		return s.insert(op.Position, op.PreviousValue)
	case *OpBatch:
		for i := len(op.Operations) - 1; i >= 0; i-- {
			if err := s.rollback(op.Operations[i]); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown operation: %T", op)
	}
}

func (s *State) insert(pos int, value int32) error {
	if pos < 0 || pos > s.array.Size() {
		return fmt.Errorf("could not insert: pos must be within bounds 0 <= %d <= %d", pos, s.array.Size())
	}
	s.array.Insert(pos, value)
	return nil
}

func (s *State) update(pos int, value int32) error {
	if pos < 0 || pos >= s.array.Size() {
		return fmt.Errorf("could not update: pos must be within bounds 0 <= %d < %d", pos, s.array.Size())
	}
	s.array.Update(pos, value)
	return nil
}

func (s *State) delete(pos int) error {
	if pos < 0 || pos >= s.array.Size() {
		return fmt.Errorf("could not delete: pos must be within bounds 0 <= %d < %d", pos, s.array.Size())
	}
	s.array.Delete(pos)
	return nil
}

func (s *State) Get(pos int) (int32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if pos < 0 || pos >= s.array.Size() {
		return 0, fmt.Errorf("could not get: pos must be non-negative and must not exceed current length")
	}
	return s.array.Get(pos), nil
}

func (s *State) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.array.Size()
}

func (s *State) Set(array []int32) {
	s.mutex.Lock()
	s.array.Set(array, 10)
	s.mutex.Unlock()
}

func (s *State) Copy() []int32 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.array.GetAll()
}
