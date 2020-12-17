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
	// TaskOutData 輸出顯示用
	TaskOutData struct {
		Which        int    // 屬於第幾個被 PoolManager 管理的 Task 物件
		Status       string `json:",omitempty"` // 執行狀態：json 用
		Caller       string `json:",omitempty"` // 上次呼叫者
		ElapseStr    string `json:",omitempty"` // 上次執行時間：json 用
		MaxCaller    string `json:",omitempty"` // 耗費時間最長的呼叫者
		MaxElapseStr string `json:",omitempty"` // 最大耗費執行時間：json 用
		Total        int64  // 總執行次數
	}

	// TaskInfo : 工作訊息，admin 查詢用
	TaskInfo struct {
		idle      bool          // 是否閒置中
		begin     time.Time     // 執行開始時間
		elapse    time.Duration // 上次執行時間
		maxElapse time.Duration // 最大耗費執行時間
		lock      *sync.RWMutex // 互斥鎖
		TaskOutData
	}
)

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// NewTaskInfo : TaskInfo object creator.
// @param	which		identity number.
func NewTaskInfo(which int) *TaskInfo {
	return &TaskInfo{
		idle: true,
		lock: new(sync.RWMutex),
		TaskOutData: TaskOutData{
			Which: which,
		},
	}
}

//------------------------------------------------------------------------------
//	Private Methods
//------------------------------------------------------------------------------

func (t *TaskInfo) prepare(caller string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.idle = false
	t.Status = "Run"
	t.begin = time.Now()
	t.Caller = caller
}

func (t *TaskInfo) completed() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.elapse = time.Since(t.begin)
	t.ElapseStr = t.elapse.String()
	if t.maxElapse < t.elapse {
		t.maxElapse = t.elapse
		t.MaxElapseStr = t.maxElapse.String()
		t.MaxCaller = t.Caller
	}
	t.Total++
	t.idle = true
	t.Status = ""
}

func (t *TaskInfo) clone() *TaskInfo {
	t.lock.Lock()
	defer t.lock.Unlock()
	res := *t
	res.lock = new(sync.RWMutex)
	return &res
}
