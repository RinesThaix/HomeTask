package state

import (
	"fmt"
	"sync"
)

type Versioner struct {
	State          *State
	transformer    *OperationalTransformer
	minVersion     int
	maxHistorySize int
	history        []Operation
	mutex          sync.RWMutex
}

var emptyHistory []Operation

func NewVersioner(state *State, maxHistorySize int) *Versioner {
	return &Versioner{
		State: state,
		transformer: &OperationalTransformer{},
		maxHistorySize: maxHistorySize,
		history: make([]Operation, 0),
		mutex: sync.RWMutex{},
	}
}

// returns: whether to rollback, diff, error
func (v *Versioner) ProcessOperation(version int, operation Operation) (bool, []Operation, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	operations, err := v.getOperationsSince(version)
	if err != nil {
		return true, nil, err
	}
	transformed := false
	for _, op := range operations {
		if res, err := v.transformer.transform(op, &operation, v.State); err != nil {
			return true, nil, err
		} else {
			transformed = transformed || res
		}
	}
	if err := v.State.Perform(operation); err != nil {
		return true, nil, err
	}
	v.newOperation(operation)
	if len(operations) == 0 {
		return false, nil, nil
	}
	operations = append(operations, operation)
	return true, operations, nil
}

func (v *Versioner) GetOperationsSince(version int) ([]Operation, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.getOperationsSince(version)
}

func (v *Versioner) getOperationsSince(version int) ([]Operation, error) {
	currentVersion := v.getCurrentVersion()
	if version == currentVersion {
		return emptyHistory, nil
	}
	if version < 0 {
		return nil, fmt.Errorf("received negative version: %d", version)
	}
	if version < v.minVersion {
		return nil, fmt.Errorf("received version %d, that is below the minimal one: %d", version, v.minVersion)
	}
	if version > currentVersion {
		return nil, fmt.Errorf("received version from the future: %d, whilst current one is %d", version, currentVersion)
	}
	//if true {
	//	return v.history[version - v.minVersion:], nil
	//}
	offset := version - v.minVersion
	result := make([]Operation, currentVersion-version)
	for i := 0; i < len(result); i++ {
		result[i] = v.history[i + offset]
	}
	return result, nil
}

func (v *Versioner) GetCurrentState() (int, []int32) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.getCurrentVersion(), v.State.Copy()
}

func (v *Versioner) getCurrentVersion() int {
	return v.minVersion + len(v.history)
}

func (v *Versioner) newOperation(op Operation) {
	if len(v.history) == v.maxHistorySize {
		v.minVersion++
		v.history = v.history[1:]
	}
	v.history = append(v.history, op)
}
