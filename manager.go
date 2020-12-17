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
	"time"
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

type (
	// ItemInterface : 存入資料的 interface
	ItemInterface interface {
		// 加入管理時呼叫一次
		OnCreate()
		// 移除管理時呼叫一次
		OnRemove()
		// 定期更新
		OnUpdate()
	}

	// Manager : 管理者物件
	Manager struct {
		datas    map[interface{}]ItemInterface
		interval time.Duration
		sync.RWMutex
	}
)

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// Init : 初始化管理者，只要使用 Manager 者，必須呼叫一次！
// @param	update		該管理者是否要定時呼叫更新 (OnUpdate)
// @param	interval	更新頻率
func (m *Manager) Init(update bool, interval time.Duration) {
	if m.datas == nil {
		m.datas = make(map[interface{}]ItemInterface)
		m.interval = interval
	} else {
		Error("Manager:Init: already init.")
		return
	}
	if update {
		go m.managerProcess()
	}
}

// Len : 內容大小
// @return	已經擁有多少內容
func (m *Manager) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.datas)
}

// Add : 新增物件至管理者中
// @param	key		標示鍵值
// @param	item	ItemInterface 物件
// @return	true 成功加入, false 失敗
func (m *Manager) Add(key interface{}, item ItemInterface) bool {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.datas[key]; ok {
		Error("Manager:Add: duplicate key. KEY=%v", key)
		return false
	}
	item.OnCreate()
	m.datas[key] = item
	return true
}

// Get : 依照鍵值取得物件
// @param	key		要取得物件的鍵值
// @return	有指定鍵值物件於管理者中時，則回傳該物件；如無，則回傳 nil
func (m *Manager) Get(key interface{}) ItemInterface {
	m.Lock()
	defer m.Unlock()
	if item, ok := m.datas[key]; ok {
		return item
	}
	return nil
}

// Remove : 依鍵值移除管理者中管理的物件
// @param	key		指定的物件鍵值
// @return	true 成功移除, false 無此物件
func (m *Manager) Remove(key interface{}) bool {
	m.Lock()
	defer m.Unlock()
	if data, ok := m.datas[key]; ok {
		data.OnRemove()
		delete(m.datas, key)
		return true
	}
	return false
}

// GetAll : 取回目前所有管理中的物件，填寫至 slice 中
// @param	ret		回填資料的 ItemInterface slice.
func (m *Manager) GetAll() (ret []ItemInterface) {
	m.Lock()
	defer m.Unlock()
	ret = make([]ItemInterface, len(m.datas))
	i := 0
	for _, v := range m.datas {
		ret[i] = v
		i++
	}
	return
}

// Exist : 以鍵值來判別物件是否處於管理中
// @param	key		要查詢的鍵值
// @return	true: 有, false: otherwise.
func (m *Manager) Exist(key interface{}) bool {
	m.Lock()
	defer m.Unlock()
	_, ok := m.datas[key]
	return ok
}

//------------------------------------------------------------------------------
//	Private Methods
//------------------------------------------------------------------------------

func (m *Manager) managerProcess() {
	for {
		<-time.After(m.interval)
		datas := m.GetAll()
		length := len(datas)
		if length < 1 {
			continue
		}
		for i := 0; i < length; i++ {
			datas[i].OnUpdate()
		}
	}
}
