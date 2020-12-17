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
	"runtime"
	"strings"
	"sync"

	"time"
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

type (
	poolManager struct {
		shutdownWorkChannel chan struct{}
		shutdownWaitGroup   sync.WaitGroup
		workChannel         chan *Task
		readyWorks          *ConcurrentQueue
		blockWorks          *ConcurrentSet
		activeWorkNums      *InterlockInt32
		maxWorkNums         int32
		initialize          *InterlockBool
		incomeWork          chan bool
		depJobs             *ConcurrentSet
		adminInfos          []*TaskInfo
	}
)

var (
	// PoolManager : 取得 poolManager 唯一物件
	PoolManager *poolManager
)

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// SendWork : 送出工作至 PoolManager 中
// @param	handler	要處理的 function
// @param	params	handler function 中所要處理的 parameters
func (p *poolManager) SendWork(handler interface{}, params ...interface{}) *Task {
	// check the handler, it must be a function.
	t := reflect.TypeOf(handler)
	if t.Kind() != reflect.Func {
		Error("PoolManager:SendWork: handler must be a function.")
		return nil
	}
	// gain barrier, if it exist.
	length := len(params)
	var b barrierBase
	if length > 0 {
		var ok bool
		if b, ok = params[length-1].(barrierBase); ok {
			length--
		}
	}
	hval := reflect.ValueOf(handler)
	hname := runtime.FuncForPC(hval.Pointer()).Name()
	strs := strings.Split(hname, "/")
	if len(strs) > 0 {
		hname = strings.Replace(strs[len(strs)-1], "-fm", "", 1)
	}
	// check the function input parameter count.
	if t.NumIn() != length {
		Error("PoolManager:SendWork: function params count not current. FUNC=%s, IN_SIZE=%d, P_SIZE=%d", hname, length, t.NumIn())
		return nil
	}
	// fill params
	e := make([]reflect.Value, length)
	for i := 0; i < length; i++ {
		e[i] = reflect.ValueOf(params[i])
	}
	// fire in the hole!
	work := &Task{
		which:    -1,
		state:    TASK_STATE_NEW,
		handler:  hval,
		name:     hname,
		elems:    e,
		checker:  b,
		complete: false,
	}
	work.submit()
	return work
}

// Start : 啟動 PoolManager
// @param	nums	這個 PoolManager 內有多少個 Task (goroutine) 等候處理工作
func (p *poolManager) Start(nums int) {
	if p.initialize.Value() {
		Error("PoolManager:Start: already start.")
		return
	}

	p.maxWorkNums = int32(nums)
	p.workChannel = make(chan *Task, nums)
	p.incomeWork = make(chan bool)
	p.adminInfos = make([]*TaskInfo, nums)
	p.initialize.True()
	p.shutdownWaitGroup.Add(nums)

	for i := 0; i < nums; i++ {
		p.adminInfos[i] = NewTaskInfo(i)
		go p.workProcess(i)
	}
	go p.mainProcess()
}

// Shutdown : 關閉此 PoolManager
func (p *poolManager) Shutdown() {
	if !p.initialize.Value() {
		Error("PoolManager:Shutdown: not start.")
		return
	}
	close(p.incomeWork)
	close(p.shutdownWorkChannel)
	// 停掉所有的獨立 goroutine
	jobs := p.depJobs.ToSlice()
	leng := len(jobs)
	for i := 0; i < leng; i++ {
		jobs[i].(*Job).Cancel()
	}
	p.shutdownWaitGroup.Wait()
	close(p.workChannel)

	Notice("PoolManager:Shutdown: finish.")
}

// AddLoopJob : 要求新增一個獨立執行的 goroutine，並交付給 PoolManager 管理
// @param	handler		callback method
// @param	interval	延遲執行週期，time.Duration format。設定為 0 則表示不延遲全速執行 (for loop)
// @param	params		parameters for callback method.
// @return	回傳實際執行工作的 essence.Job 物件
func (p *poolManager) AddLoopJob(handler interface{}, interval time.Duration, params ...interface{}) *Job {
	// check the handler, it must be a function.
	t := reflect.TypeOf(handler)
	if t.Kind() != reflect.Func {
		Error("PoolManager:AddLoopJob: handler must be a function.")
		return nil
	}
	length := len(params)
	hval := reflect.ValueOf(handler)
	hname := runtime.FuncForPC(hval.Pointer()).Name()
	strs := strings.Split(hname, "/")
	if len(strs) > 0 {
		hname = strings.Replace(strs[len(strs)-1], ")-fm", "", 1)
	}

	// check the function input parameter count.
	if t.NumIn() != length {
		Error("PoolManager:AddLoopJob: function params count not current. FUNC=%s, IN_SIZE=%d, P_SIZE=%d", hname, length, t.NumIn())
		return nil
	}
	// fill params
	e := make([]reflect.Value, length)
	for i := 0; i < length; i++ {
		e[i] = reflect.ValueOf(params[i])
	}
	// fire in the hole!
	job := &Job{
		state:      JOB_STATE_IDLE,
		handler:    hval,
		name:       hname,
		interval:   interval,
		elems:      e,
		resumed:    make(chan bool),
		waitGroup:  &p.shutdownWaitGroup,
		delayStart: 0,
	}
	p.depJobs.Add(job)
	return job
}

// GetAdminInfos : 取得 PoolManager 所管理的 goroutines 目前狀態
// @return TaskInfo slice.
func (p *poolManager) GetAdminInfos() []TaskInfo {
	out := make([]TaskInfo, p.maxWorkNums)
	for i := int32(0); i < p.maxWorkNums; i++ {
		out[i] = *(p.adminInfos[i].clone())
	}
	return out
}

//------------------------------------------------------------------------------
//	Private Methods
//------------------------------------------------------------------------------

func (p *poolManager) addReadyWork(work *Task) {
	// if IsServerDown() {
	// 	return
	// }
	p.readyWorks.Push(work)
	p.incomeWork <- true
}

func (p *poolManager) addBlockWork(work *Task) {
	p.blockWorks.Add(work)
}

func (p *poolManager) moveWorkToReady(work *Task) {
	// if IsServerDown() {
	// 	return
	// }
	p.blockWorks.Remove(work)
	p.readyWorks.Push(work)
	p.incomeWork <- true
}

func (p *poolManager) removeWorkFromBlock(work *Task) {
	if work.state == TASK_STATE_CANCEL {
		p.blockWorks.Remove(work)
	}
}

func (p *poolManager) workProcess(which int) {
	for {
		select {
		case <-p.shutdownWorkChannel:
			p.shutdownWaitGroup.Done()
			return

		case work := <-p.workChannel:
			p.activeWorkNums.Increment()
			work.invoke(which, p.adminInfos[which])
			p.activeWorkNums.Decrement()
		}
	}
}

func (p *poolManager) mainProcess() {
	for {
		select {
		case <-p.incomeWork:
		case <-time.After(time.Millisecond * 1):
			if p.activeWorkNums.Value() < p.maxWorkNums && !p.readyWorks.Empty() {
				p.workChannel <- p.readyWorks.Pop().(*Task)
			}
		}
	}
}

func (p *poolManager) removeJob(job *Job) {
	p.depJobs.Remove(job)
}

func (p *poolManager) catchPanic(funcName string) {
	if r := recover(); r != nil {
		buf := make([]byte, 10000)
		runtime.Stack(buf, false)
		Critical("PANIC Defered [%v] : Stack Trace : %v", r, string(buf))
	}
}

//------------------------------------------------------------------------------
//	Auto initialize function
//------------------------------------------------------------------------------

func init() {
	PoolManager = &poolManager{
		shutdownWorkChannel: make(chan struct{}),
		readyWorks:          NewConcurrentQueue(),
		blockWorks:          NewConcurrentSet(),
		activeWorkNums:      NewInterlockInt32(0),
		maxWorkNums:         0,
		initialize:          NewInterlockBool(false),
		depJobs:             NewConcurrentSet(),
	}
}
