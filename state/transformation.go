package state

import "fmt"

type OperationalTransformer struct {}

func (ot *OperationalTransformer) transform(committed Operation, transformable *Operation, state *State) (bool, error) {
	switch c := committed.(type) {
	case *OpInsert:
		return ot._transform(transformable, state, c.Position, 1)
	case *OpUpdate:
		return false, nil
	case *OpDelete:
		return ot._transform(transformable, state, c.Position, -1)
	case *OpBatch:
		result := false
		for _, op := range c.Operations {
			if res, err := ot.transform(op, transformable, state); err != nil {
				return false, err
			} else {
				result = result || res
			}
		}
		return result, nil
	default:
		return false, fmt.Errorf("unknown operation: %T", c)
	}
}

func (ot *OperationalTransformer) _transform(operation *Operation, state *State, pos, delta int) (bool, error) {
	switch o := (*operation).(type) {
	case *OpInsert:
		if o.Position >= pos {
			o.Position += delta
			if o.Position < 0 {
				o.Position = 0
			} else if o.Position > state.array.Size() {
				o.Position = state.array.Size()
			}
			return true, nil
		}
	case *OpUpdate:
		if o.Position >= pos {
			o.Position += delta
			if o.Position < 0 {
				o.Position = 0
			} else if o.Position >= state.array.Size() {
				o.Position = state.array.Size() - 1
			}
			return true, nil
		}
	case *OpDelete:
		if o.Position >= pos {
			o.Position += delta
			if o.Position < 0 {
				o.Position = 0
			} else if o.Position >= state.array.Size() {
				o.Position = state.array.Size() - 1
			}
			return true, nil
		}
	case *OpBatch:
		result := false
		for _, op := range o.Operations {
			if res, err := ot._transform(&op, state, pos, delta); err != nil {
				return false, err
			} else {
				result = result || res
			}
		}
		return result, nil
	default:
		return false, fmt.Errorf("unknown operation: %T", o)
	}
	return false, nil
}
