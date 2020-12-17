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
	"errors"
	"fmt"
	"reflect"
	"time"
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

type (
	// OrderData : 用來處理 workpool 中 "依序處理" 所需要的資料結構
	// 要依序處理者，必須將 OrderData embedded 到自己的物件中，如同 essence.Player
	OrderData struct {
		works  *ConcurrentQueue
		delays *ConcurrentSet
		tag    string
	}

	barrierBase interface {
		isClear(work *Task) bool

		setup(work *Task)

		cancel(work *Task)

		completed(work *Task)
	}

	barrier struct {
		data *OrderData
	}

	multiBarrier struct {
		datas []*OrderData
	}

	delayBarrier struct {
		data     *OrderData
		work     *Task
		duration time.Duration
		timer    *time.Timer
		expired  bool
	}

	delayMultiBarrier struct {
		datas    []*OrderData
		work     *Task
		duration time.Duration
		timer    *time.Timer
		expired  bool
	}
)

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// MakeBarrier : Make barrier object to keep worker in sequence.
// @param	params	要拿來排隊的物件，物件必須 embedded essence.OrderData 才能處理
func MakeBarrier(params ...interface{}) interface{} {
	if err := orderChecker(params); err != nil {
		Error("MakeBarrier: ERR=%s", err.Error())
		return nil
	}
	length := len(params)
	if length == 1 {
		order := reflect.ValueOf(params[0]).Elem().FieldByName("OrderData").Interface().(OrderData)
		b := &barrier{
			data: order.GetOrder(),
		}
		return b
	}
	b := &multiBarrier{
		datas: make([]*OrderData, length),
	}
	for i := 0; i < length; i++ {
		order := reflect.ValueOf(params[i]).Elem().FieldByName("OrderData").Interface().(OrderData)
		b.datas[i] = order.GetOrder()
	}
	return b
}

// DelayBarrier : Make barrier object to keep worker in sequence.
// @param	duration	要延遲多久執行
// @param	params		要拿來排隊的物件，物件必須 embedded essence.OrderData 才能處理
func DelayBarrier(duration time.Duration, params ...interface{}) interface{} {
	if err := orderChecker(params); err != nil {
		Error("DelayBarrier: ERR=%s", err.Error())
		return nil
	}
	length := len(params)
	if length == 1 {
		order := reflect.ValueOf(params[0]).Elem().FieldByName("OrderData").Interface().(OrderData)
		b := &delayBarrier{
			data:     order.GetOrder(),
			work:     nil,
			duration: duration,
			timer:    nil,
			expired:  false,
		}
		return b
	}
	b := &delayMultiBarrier{
		datas:    make([]*OrderData, length),
		work:     nil,
		duration: duration,
		timer:    nil,
		expired:  false,
	}
	for i := 0; i < length; i++ {
		order := reflect.ValueOf(params[i]).Elem().FieldByName("OrderData").Interface().(OrderData)
		b.datas[i] = order.GetOrder()
	}
	return b
}

// GetOrder : 取回自身 ptr, this method just use for MakeBarrier().
func (o *OrderData) GetOrder() *OrderData {
	return o
}

// OrderInit : 初始 OrderData, embedded OrderData 的物件必須在建立物件時呼叫一次
func (o *OrderData) OrderInit(msg string) {
	//	o.first = nil
	o.works = NewConcurrentQueue()
	o.delays = NewConcurrentSet()
	o.tag = msg
}

//------------------------------------------------------------------------------
//	Private Methods
//------------------------------------------------------------------------------

func orderChecker(params []interface{}) error {
	if params == nil {
		return errors.New("nil params")
	}
	length := len(params)
	if length < 1 {
		return errors.New("empty params")
	}
	// Check params type
	for i := 0; i < length; i++ {
		t := reflect.TypeOf(params[i])
		if t.Kind() != reflect.Ptr {
			return fmt.Errorf("[%d] param must be a ptr", i+1)
		}
		if !reflect.ValueOf(params[i]).Elem().FieldByName("OrderData").IsValid() {
			return fmt.Errorf("[%d] param must have 'OrderData' embedded field", i+1)
		}
	}
	return nil
}

//------------------------------------------------------------------------------

func (o *OrderData) getFirstWork() *Task {
	if first := o.works.Peek(); first != nil {
		return first.(*Task)
	}
	return nil
}

func (o *OrderData) addWork(work *Task) {
	o.works.Push(work)
}

func (o *OrderData) addDelayWork(work *Task) {
	o.delays.Add(work)
}

func (o *OrderData) completeWork(work *Task) {
	o.removeFirstWork(work)
}

func (o *OrderData) removeFirstWork(work *Task) {
	if o.getFirstWork() != work {
		Error("OrderData:removeFirstWork: diff work. NAME=%s", work.name)
		return
	}
	o.works.Pop()
	if !o.works.Empty() {
		o.works.Peek().(*Task).reinvoke()
	}
}

func (o *OrderData) cancelFirstWork(work *Task) {
	if o.getFirstWork() == work {
		o.removeFirstWork(work)
	}
}

func (o *OrderData) expireDelayed(work *Task) {
	if o.delays.Contains(work) == false {
		Error("OrderData:expireDelayed: not found. NAME=%s", work.name)
		return
	}
	o.delays.Remove(work)
	if work.state == TASK_STATE_CANCEL {
		return
	}
	o.addWork(work)
}

//------------------------------------------------------------------------------
//	barrier
//------------------------------------------------------------------------------

//------------------------------------------------------------------------------

func (b *barrier) isClear(work *Task) bool {
	return work == b.data.getFirstWork()
}

func (b *barrier) setup(work *Task) {
	b.data.addWork(work)
}

func (b *barrier) cancel(work *Task) {
	b.data.cancelFirstWork(work)
}

func (b *barrier) completed(work *Task) {
	b.data.completeWork(work)
}

//------------------------------------------------------------------------------

func (m *multiBarrier) isClear(work *Task) bool {
	length := len(m.datas)
	for i := 0; i < length; i++ {
		if work != m.datas[i].getFirstWork() {
			return false
		}
	}
	return true
}

func (m *multiBarrier) setup(work *Task) {
	length := len(m.datas)
	for i := 0; i < length; i++ {
		m.datas[i].addWork(work)
	}
}

func (m *multiBarrier) cancel(work *Task) {
	length := len(m.datas)
	for i := 0; i < length; i++ {
		m.datas[i].cancelFirstWork(work)
	}
}

func (m *multiBarrier) completed(work *Task) {
	length := len(m.datas)
	for i := 0; i < length; i++ {
		m.datas[i].completeWork(work)
	}
}

//------------------------------------------------------------------------------

func (d *delayBarrier) isClear(work *Task) bool {
	if d.expired == false {
		return false
	}
	return d.data.getFirstWork() == work
}

func (d *delayBarrier) setup(work *Task) {
	d.work = work
	d.timer = time.AfterFunc(d.duration, d.onExpired)
	d.data.addDelayWork(work)
}

func (d *delayBarrier) cancel(work *Task) {
	if d.timer == nil {
		return
	}
	d.timer.Stop()
	if d.expired == false {
		d.data.expireDelayed(work)
	}
	d.data.cancelFirstWork(work)
}

func (d *delayBarrier) completed(work *Task) {
	d.data.completeWork(work)
}

func (d *delayBarrier) onExpired() {
	if d.expired == false {
		d.expired = true
		d.data.expireDelayed(d.work)
	}
	d.work.reinvoke()
}

//------------------------------------------------------------------------------

func (m *delayMultiBarrier) isClear(work *Task) bool {
	if m.expired == false {
		return false
	}
	length := len(m.datas)
	for i := 0; i < length; i++ {
		if m.datas[i].getFirstWork() != work {
			return false
		}
	}
	return true
}

func (m *delayMultiBarrier) setup(work *Task) {
	m.work = work
	m.timer = time.AfterFunc(m.duration, m.onExpired)
	length := len(m.datas)
	for i := 0; i < length; i++ {
		m.datas[i].addDelayWork(work)
	}
}

func (m *delayMultiBarrier) cancel(work *Task) {
	if m.timer == nil {
		return
	}
	length := len(m.datas)
	m.timer.Stop()
	if m.expired == false {
		for i := 0; i < length; i++ {
			m.datas[i].expireDelayed(work)
		}
	}
	for i := 0; i < length; i++ {
		m.datas[i].cancelFirstWork(work)
	}
}

func (m *delayMultiBarrier) completed(work *Task) {
	length := len(m.datas)
	for i := 0; i < length; i++ {
		m.datas[i].completeWork(work)
	}
}

func (m *delayMultiBarrier) onExpired() {
	if m.expired == false {
		m.expired = true
		length := len(m.datas)
		for i := 0; i < length; i++ {
			m.datas[i].expireDelayed(m.work)
		}
	}
	m.work.reinvoke()
}
