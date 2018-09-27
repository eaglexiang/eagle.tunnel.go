package eaglelib

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

// Push 在链表尾部增加节点
func (l *SyncList) Push(v interface{}) {
	l.lock.Lock()
	l.raw.PushBack(v)
	l.lock.Unlock()
}

// Remove 删除指定节点
func (l *SyncList) Remove(e *list.Element) {
	l.lock.Lock()
	l.raw.Remove(e)
	l.lock.Unlock()
}

// Front 返回链表的头节点
func (l *SyncList) Front() *list.Element {
	return l.raw.Front()
}
