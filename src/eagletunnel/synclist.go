package eagletunnel

import (
	"container/list"
	"sync"
)

type SyncList struct {
	raw  *list.List
	lock sync.Mutex
}

func NewSyncList() *SyncList {
	result := SyncList{
		raw: list.New()}
	return &result
}

func (l *SyncList) push(v interface{}) {
	l.lock.Lock()
	l.raw.PushBack(v)
	l.lock.Unlock()
}

func (l *SyncList) remove(e *list.Element) {
	l.lock.Lock()
	l.raw.Remove(e)
	l.lock.Unlock()
}
