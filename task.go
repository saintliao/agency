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
	"reflect"
	"strconv"
	"sync"
)

//------------------------------------------------------------------------------
// Enumeration
//------------------------------------------------------------------------------

// TaskStateEnum : Task 狀態
type TaskStateEnum int

const (
	// TaskStateNew : 新狀態
	TaskStateNew TaskStateEnum = iota
	// TaskStateBlocked : 狀態阻斷中
	TaskStateBlocked
	// TaskStateReady : 預備中，事務可以被處裡
	TaskStateReady
	// TaskStateCancel : 取消
	TaskStateCancel
	// TaskStateInvoked : 事務處理中
	TaskStateInvoked
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

type (
	// Task : 由 PoolManager 所管理的 goroutine 包裝，用來仿造 ThreadPool 內的個
	// 別 Thread 使用
	Task struct {
		which    int             // 屬於第幾個被 PoolManager 管理的 Task 物件
		state    TaskStateEnum   // 目前狀態
		handler  reflect.Value   // 處理事務的 function
		name     string          // function name
		elems    []reflect.Value // function parameters
		checker  barrierBase     // 排隊用物件
		complete bool            // 確認事務是否已處理完成
		sync.Mutex
	}
)

//------------------------------------------------------------------------------
//	Variables
//------------------------------------------------------------------------------

var (
	taskEnumStrinMap = map[TaskStateEnum]string{
		TaskStateNew:     "TaskStateNew",
		TaskStateBlocked: "TaskStateBlocked",
		TaskStateReady:   "TaskStateReady",
		TaskStateCancel:  "TaskStateCancel",
		TaskStateInvoked: "TaskStateInvoked",
	}
)

//------------------------------------------------------------------------------
// Public Methods
//------------------------------------------------------------------------------

func (t *TaskStateEnum) String() string {
	return taskEnumStrinMap[*t]
}

// Cancel : 取消命令，多半都是用在 delayBarrier 上
func (w *Task) Cancel() {
	if TaskStateReady < w.state {
		return
	}
	w.Lock()
	defer w.Unlock()
	switch w.state {
	case TaskStateBlocked, TaskStateReady:
		old := w.state
		w.state = TaskStateCancel
		if old == TaskStateBlocked {
			PoolManager.removeWorkFromBlock(w)
		}
		if w.checker != nil {
			w.checker.cancel(w)
		}
		Info("Task:Cancel: NAME=%s", w.name)
	}
}

//------------------------------------------------------------------------------
//	Private Methods
//------------------------------------------------------------------------------

func (w *Task) submit() {
	if w.state != TaskStateNew {
		Error("Task:submit: failed. STAT=%s", strconv.Itoa(int(w.state)))
		return
	}
	if w.checker != nil {
		w.checker.setup(w)
	}
	w.Lock()
	defer w.Unlock()
	if w.canInvoke() {
		w.state = TaskStateReady
		PoolManager.addReadyWork(w)
	} else {
		w.state = TaskStateBlocked
		PoolManager.addBlockWork(w)
	}
}

func (w *Task) canInvoke() bool {
	if w.checker == nil {
		return true
	}
	return w.checker.isClear(w)
}

func (w *Task) isCompleted() bool {
	return w.complete
}

func (w *Task) completed() {
	if w.state != TaskStateInvoked {
		Error("Task:completed: wrong state. STATE=%s", w.state.String())
		return
	}
	w.complete = true
	if w.checker != nil {
		w.checker.completed(w)
	}
}

func (w *Task) invoke(which int, info *TaskInfo) {
	if w.state == TaskStateCancel {
		return
	}
	w.Lock()
	if w.state != TaskStateReady {
		Error("Task:invoke: wrong state. STATE=%s", w.state.String())
		w.Unlock()
		return
	}
	w.which = which
	w.state = TaskStateInvoked
	w.Unlock()
	info.prepare(w.name)
	defer PoolManager.catchPanic(w.name)
	w.handler.Call(w.elems)
	info.completed()
	w.completed()
}

func (w *Task) reinvoke() {
	if TaskStateBlocked < w.state {
		return
	}
	w.Lock()
	defer w.Unlock()
	switch w.state {
	case TaskStateBlocked:
		if w.canInvoke() {
			w.state = TaskStateReady
			PoolManager.moveWorkToReady(w)
		}
	case TaskStateReady, TaskStateCancel, TaskStateInvoked:
		return
	}
}
