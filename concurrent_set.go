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

// ConcurrentSet 簡易的 parallel set container.
type ConcurrentSet struct {
	sync.RWMutex
	items map[interface{}]interface{}
}

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// NewConcurrentSet 取得一個新的 ConcurrentSet 物件
func NewConcurrentSet() *ConcurrentSet {
	set := &ConcurrentSet{}
	set.items = make(map[interface{}]interface{})
	return set
}

// Len 取得此 container 的大小
func (s *ConcurrentSet) Len() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.items)
}

// Empty 測試此 container 是否為空
func (s *ConcurrentSet) Empty() bool {
	s.RLock()
	defer s.RUnlock()
	return len(s.items) == 0
}

// ToSlice 將資料轉換為 slice
func (s *ConcurrentSet) ToSlice() []interface{} {
	s.Lock()
	defer s.Unlock()
	result := make([]interface{}, len(s.items))
	i := 0
	for k := range s.items {
		result[i] = k
		i++
	}
	return result
}

// Add 加入 set 的鍵值
func (s *ConcurrentSet) Add(obj interface{}) {
	s.Lock()
	defer s.Unlock()
	s.items[obj] = true
}

// Contains 測試鍵值是否存在
func (s *ConcurrentSet) Contains(obj interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	return s.items[obj] != nil
}

// Remove 移除鍵值，return true: successfully, false: otherwise.
func (s *ConcurrentSet) Remove(obj interface{}) bool {
	s.Lock()
	defer s.Unlock()

	if found := s.items[obj]; found != nil {
		delete(s.items, obj)
		return true
	}
	return false
}

// Clear all container.
func (s *ConcurrentSet) Clear() {
	s.Lock()
	defer s.Unlock()
	s.items = make(map[interface{}]interface{})
}
