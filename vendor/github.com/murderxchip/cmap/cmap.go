package cmap

import (
	"sync"
)

type CMap struct {
	v        map[string]interface{}
	l        sync.RWMutex
	afterSet CbAfterSet
	afterGet CbAfterGet
}

type MapItem struct {
	Key   string
	Value interface{}
}

type CbAfterSet func()
type CbAfterGet func()

func NewCMap() *CMap {
	return &CMap{v: make(map[string]interface{})}
}

func (m *CMap) SetListenerSet(listener CbAfterSet) {
	m.afterSet = listener
}

func (m *CMap) SetListenerGet(listener CbAfterGet) {
	m.afterGet = listener
}

func (m *CMap) Size() int {
	return len(m.v)
}

func (m *CMap) Set(key string, value interface{}) error {
	m.l.Lock()
	defer m.l.Unlock()
	m.v[key] = value
	if m.afterSet != nil {
		m.afterSet()
	}

	return nil
}

func (m *CMap) Get(key string) (value interface{}, exists bool) {
	m.l.RLock()
	defer m.l.RUnlock()
	value, exists = m.v[key]
	return
}

func (m *CMap) Exists(key string) bool {
	m.l.RLock()
	defer m.l.RUnlock()
	_, exists := m.v[key]
	return exists
}

func (m *CMap) Dump() <-chan MapItem {
	outChan := make(chan MapItem, m.Size())
	for k, v := range m.v {
		outChan <- MapItem{
			Key:   k,
			Value: v,
		}
	}
	close(outChan)
	return outChan
}
