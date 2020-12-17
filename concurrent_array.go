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

// ConcurrentArray parallel array container with correntency.
type ConcurrentArray struct {
	sync.RWMutex
	items []interface{}
	count int
}

//------------------------------------------------------------------------------
//	Methods: Array
//------------------------------------------------------------------------------

// NewConcurrentArray Retrieves a new corrent array container object.
func NewConcurrentArray() *ConcurrentArray {
	arr := &ConcurrentArray{}
	arr.items = make([]interface{}, 50)
	arr.count = 0
	return arr
}

// Len Gains the legth of this corrent array.
func (a *ConcurrentArray) Len() int {
	a.Lock()
	defer a.Unlock()
	return a.count
}

// Empty Tests this array is empty or not.
func (a *ConcurrentArray) Empty() bool {
	return a.Len() == 0
}

// Append the given item into array.
func (a *ConcurrentArray) Append(item interface{}) {
	a.Lock()
	defer a.Unlock()
	a.add(item)
}

// Appends given items into array.
func (a *ConcurrentArray) Appends(items ...interface{}) {
	length := len(items)
	a.Lock()
	defer a.Unlock()
	for i := 0; i < length; i++ {
		a.add(items[i])
	}
}

// Remove the given item element from array.
func (a *ConcurrentArray) Remove(item interface{}) bool {
	a.Lock()
	defer a.Unlock()
	index := a.indexOf(item)
	if index == -1 {
		return false
	}
	a.items[index] = nil
	for i := index; i < a.count-1; i++ {
		a.items[i], a.items[i+1] = a.items[i+1], a.items[i]
	}
	a.count--
	return true
}

// Get elemet for given index, index should beginning from zero. If the index is
// out of the range, will return nil.
func (a *ConcurrentArray) Get(index int) interface{} {
	a.Lock()
	defer a.Unlock()
	if index < 0 || index >= a.count {
		return nil
	}
	return a.items[index]
}

// IndexOf  Get the array index for the given item.
func (a *ConcurrentArray) IndexOf(item interface{}) int {
	a.Lock()
	defer a.Unlock()
	return a.indexOf(item)
}

// Contains  Tests the given item has inside this array.
func (a *ConcurrentArray) Contains(item interface{}) bool {
	a.Lock()
	defer a.Unlock()
	return a.indexOf(item) != -1
}

// ToSlice  Convert all array items to slice and return.
func (a *ConcurrentArray) ToSlice() []interface{} {
	a.Lock()
	defer a.Unlock()
	res := make([]interface{}, a.count)
	copy(res, a.items)
	return res
}

// GainFrom  Given the begin index and item nums, will retrun the given range elements.
func (a *ConcurrentArray) GainFrom(index int, nums int) []interface{} {
	a.Lock()
	defer a.Unlock()
	if index < 0 || index >= a.count {
		return nil
	}
	length := a.count - index
	if length > nums {
		length = nums
	}
	res := make([]interface{}, length)
	copy(res, a.items[index:])
	return res
}

// Clear this array.
func (a *ConcurrentArray) Clear() {
	a.Lock()
	defer a.Unlock()
	a.items = make([]interface{}, 50)
	a.count = 0
}

func (a *ConcurrentArray) add(item interface{}) {
	a.items[a.count] = item
	a.count++
	a.reduceSize()
}

func (a *ConcurrentArray) reduceSize() {
	capa := cap(a.items)
	if a.count >= (capa - 1) {
		newCapa := (capa + 1) * 2
		temp := make([]interface{}, newCapa, newCapa)
		copy(temp, a.items)
		a.items = temp
	}
}

func (a *ConcurrentArray) indexOf(item interface{}) int {
	index := -1
	for i := 0; i < a.count; i++ {
		if a.items[i] == item {
			index = i
			break
		}
	}
	return index
}
