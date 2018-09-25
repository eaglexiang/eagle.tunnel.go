package eagletunnel

import (
	"container/list"
	"sync"
)

// SyncList 线程/协程 安全的同步链表
type SyncList struct {
	raw  *list.List
	lock sync.Mutex
}

// CreateSyncList 创建新的SyncList
func CreateSyncList() *SyncList {
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
