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

//	Structures declare

type (
	// ConcurrentMap 簡易的 parallel map container.
	ConcurrentMap struct {
		sync.RWMutex
		datas map[interface{}]interface{}
	}

	// PairValue 給 Snapshot 用的 Pair 物件
	PairValue struct {
		Key   interface{}
		Value interface{}
	}
)

//------------------------------------------------------------------------------
//	Methods: Array
//------------------------------------------------------------------------------

// NewConcurrentMap 取得一個新的 ConcurrentMap 物件
func NewConcurrentMap() *ConcurrentMap {
	data := &ConcurrentMap{
		datas: make(map[interface{}]interface{}),
	}
	return data
}

// Len 取得 ConcurrentMap 的大小
func (m *ConcurrentMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.datas)
}

// Empty 測試此 ConcurrentMap 是否為空
func (m *ConcurrentMap) Empty() bool {
	m.RLock()
	defer m.RUnlock()
	return len(m.datas) != 0
}

// Set 以 key-value 格式填入資料
func (m *ConcurrentMap) Set(key interface{}, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.datas[key] = value
}

// Get 以 key 值取得 value，如無此 key 則回傳 nil
func (m *ConcurrentMap) Get(key interface{}) interface{} {
	m.RLock()
	defer m.RUnlock()
	if v, ok := m.datas[key]; ok {
		return v
	}
	return nil
}

// Remove 以 key 值移除資料
func (m *ConcurrentMap) Remove(key interface{}) {
	m.Lock()
	defer m.Unlock()
	delete(m.datas, key)
}

// Exist 以 key 值測試資料是否存在
func (m *ConcurrentMap) Exist(key interface{}) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.datas[key]
	return ok
}

// Clear 清除整個 ConcurrentMap 資料
func (m *ConcurrentMap) Clear() {
	m.Lock()
	defer m.Unlock()
	m.datas = make(map[interface{}]interface{})
}

// GetSnapshot 取得當下資料的 snapshot，將會回傳 PairValue 的 slice.
func (m *ConcurrentMap) GetSnapshot() []*PairValue {
	m.Lock()
	defer m.Unlock()
	ret := make([]*PairValue, len(m.datas))
	i := 0
	for k, v := range m.datas {
		ret[i] = &PairValue{k, v}
		i++
	}
	return ret
}
