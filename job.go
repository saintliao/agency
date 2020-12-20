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

import (
	"reflect"
	"sync"
	"time"
)

//------------------------------------------------------------------------------
// Enumeration
//------------------------------------------------------------------------------

// JobStateEnum Enumeration for job state.
type JobStateEnum int

const (
	// JobStateIdle : 初始狀態
	JobStateIdle JobStateEnum = iota
	// JobStateRun : 執行中
	JobStateRun
	// JobStateSuspend : 暫停中
	JobStateSuspend
	// JobStateCancel : 停止
	JobStateCancel
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

// Job 獨立工作
type Job struct {
	handler    reflect.Value
	name       string
	elems      []reflect.Value
	interval   time.Duration
	state      JobStateEnum
	resumed    chan bool
	waitGroup  *sync.WaitGroup
	delayStart time.Duration
}

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// Run : 開始執行工作 (必須呼叫！)
// @param	delay	延遲多久後開始工作 time.Duration 格式，可不輸入
func (j *Job) Run(delay ...interface{}) {
	if j.state != JobStateIdle {
		Error("Job:Run: invalid state. FUNC=%s, STATE=%d", j.name, j.state)
		return
	}
	j.waitGroup.Add(1)
	if delay != nil {
		ok := false
		if j.delayStart, ok = delay[0].(time.Duration); !ok {
			j.delayStart = 0
		}
	}
	j.state = JobStateRun
	go j.jobProcess()
}

// Suspend : 暫停執行工作
func (j *Job) Suspend() {
	if j.state == JobStateSuspend {
		return
	}
	j.state = JobStateSuspend
}

// Resume : 恢復執行工作
func (j *Job) Resume() {
	if j.state != JobStateSuspend {
		return
	}
	j.state = JobStateRun
	j.resumed <- true
}

// Cancel : 結束工作
func (j *Job) Cancel() {
	j.state = JobStateCancel
}

// GetStatus : 取得目前工作狀態
func (j *Job) GetStatus() JobStateEnum {
	return j.state
}

//------------------------------------------------------------------------------
// Private Methods
//------------------------------------------------------------------------------

func (j *Job) jobProcess() {
	if j.delayStart > 0 {
		<-time.After(j.delayStart)
	}
	for {
		switch j.state {
		case JobStateRun:
			j.handler.Call(j.elems)

		case JobStateSuspend:
			<-j.resumed

		case JobStateCancel:
			j.waitGroup.Done()
			PoolManager.removeJob(j)
			Info("Job:process: job end. NAME=%s", j.name)
			return
		}
		// call for delay.
		if j.interval > 0 {
			<-time.After(j.interval)
		}
	}
}
