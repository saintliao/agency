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
	"sync/atomic"
)

//	Type Declare

type (
	// InterlockBool : 仿 Gt2 同名玩意
	InterlockBool int32

	// InterlockInt32 : 仿 Gt2 同名玩意
	InterlockInt32 int32

	// InterlockInt64 : 仿 Gt2 同名玩意
	InterlockInt64 int64
)

//------------------------------------------------------------------------------
//	Methods
//------------------------------------------------------------------------------

// NewInterlockBool : 取回 InterlockBool 物件指標
// @param	ok	預設值
// @return	InterlockBool 物件指標
func NewInterlockBool(ok bool) *InterlockBool {
	data := new(InterlockBool)
	if ok {
		data.True()
	}
	return data
}

// NewInterlockInt32 : 取回 InterlockInt32 物件指標
// @param	val	預設值
// @return	InterlockInt32 物件指標
func NewInterlockInt32(val int32) *InterlockInt32 {
	data := new(InterlockInt32)
	atomic.StoreInt32((*int32)(data), val)
	return data
}

// NewInterlockInt64 : 取回 InterlockInt64 物件指標
// @param	val	預設值
// @return	InterlockInt64 物件指標
func NewInterlockInt64(val int64) *InterlockInt64 {
	data := new(InterlockInt64)
	atomic.StoreInt64((*int64)(data), val)
	return data
}

// InterlockBool

// True : 指定 InterlockBool 物件為 true
func (b *InterlockBool) True() {
	atomic.StoreInt32((*int32)(b), 1)
}

// False : 指定 InterlockBool 為 false
func (b *InterlockBool) False() {
	atomic.StoreInt32((*int32)(b), 0)
}

// Value : 取出 InterlockBool 值
func (b *InterlockBool) Value() bool {
	return atomic.LoadInt32((*int32)(b)) == 1
}

// Exchange : 交換資料, 將會回傳舊值
func (b *InterlockBool) Exchange(new bool) bool {
	var val int32
	if new {
		val = 1
	}
	return atomic.SwapInt32((*int32)(b), val) == 1
}

// InterlockInt32

// Value : 取出 InterlockInt32 值
func (i *InterlockInt32) Value() int32 {
	return atomic.LoadInt32((*int32)(i))
}

// Increment : increase InterlockInt32 value 1
func (i *InterlockInt32) Increment() int32 {
	return atomic.AddInt32((*int32)(i), 1)
}

// Decrement : decrease InterlockInt32 value 1
func (i *InterlockInt32) Decrement() int32 {
	return atomic.AddInt32((*int32)(i), -1)
}

// Exchange : 交換資料，將會回傳舊值
func (i *InterlockInt32) Exchange(new int32) int32 {
	return atomic.SwapInt32((*int32)(i), new)
}

// InterlockInt64

// Value : 取出 InterlockInt64 值
func (i *InterlockInt64) Value() int64 {
	return atomic.LoadInt64((*int64)(i))
}

// Increment : increase InterlockInt64 value 1
func (i *InterlockInt64) Increment() int64 {
	return atomic.AddInt64((*int64)(i), 1)
}

// Decrement : decrease InterlockInt64 value 1
func (i *InterlockInt64) Decrement() int64 {
	return atomic.AddInt64((*int64)(i), -1)
}

// Exchange : 交換資料，將會回傳舊值
func (i *InterlockInt64) Exchange(new int64) int64 {
	return atomic.SwapInt64((*int64)(i), new)
}
