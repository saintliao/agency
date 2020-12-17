//------------------------------------------------------------------------------
//
//  Copyright 2020 by International Games System Co., Ltd.
//  All rights reserved.
//
//  This software is the confidential and proprietary information of
//  International Game System Co., Ltd. ('Confidential Information'). You shall
//  not disclose such Confidential Information and shall use it only in
//  accordance with the terms of the license agreement you entered into with
//  International Game System Co., Ltd.
//
//------------------------------------------------------------------------------

//------------------------------------------------------------------------------
//	Package declare
//------------------------------------------------------------------------------

package agency

//------------------------------------------------------------------------------
//	Import packages
//------------------------------------------------------------------------------

import (
	"sync"
)

//	Structure declare

type (
	node struct {
		data interface{}
		next *node
	}

	// ConcurrentQueue 簡單的 concurrent queue container.
	ConcurrentQueue struct {
		sync.RWMutex
		head  *node
		tail  *node
		count int
	}
)

//------------------------------------------------------------------------------
//	Methods
//------------------------------------------------------------------------------

// NewConcurrentQueue 取得新 ConcurrentQueue 物件
func NewConcurrentQueue() *ConcurrentQueue {
	q := &ConcurrentQueue{
		count: 0,
	}
	return q
}

// Empty 檢查 ConcurrentQueue 是否為 empty container.
// @return	true: yep, false: otherwise.
func (q *ConcurrentQueue) Empty() bool {
	q.Lock()
	defer q.Unlock()
	return q.count < 1
}

// Len Retrieves the item nums of this queue contain.
func (q *ConcurrentQueue) Len() int {
	q.Lock()
	defer q.Unlock()
	return q.count
}

// Push Add item object into this queue.
// @param	item	item object, with interface type.
func (q *ConcurrentQueue) Push(item interface{}) {
	q.Lock()
	defer q.Unlock()

	n := &node{data: item}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.count++
}

// Pop Gain the front item from queue.
// @return	front item object from queue.
func (q *ConcurrentQueue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	if q.head == nil {
		return nil
	}

	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--

	return n.data
}

// Peek the front item in queue, but not pop it.
// @return	front item object in the queue.
func (q *ConcurrentQueue) Peek() interface{} {
	q.Lock()
	defer q.Unlock()

	n := q.head
	if n != nil {
		return n.data
	}
	return nil
}
