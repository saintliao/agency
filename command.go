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
	"encoding/binary"
	"errors"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

// Command : 通訊協定封包
type Command struct {
	cmd    uint32 // 命令型別
	length uint32 // body 長度
	body   []byte // 命令資料
}

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

// CreateCommand : 以 byte array 建立 Command 物件
// @param	data	byte array
// @return	Command object & error
func CreateCommand(data []byte) (*Command, error) {
	length := len(data)
	if length < 8 {
		return nil, errors.New("invalid length")
	}
	return &Command{binary.LittleEndian.Uint32(data[0:4]), binary.LittleEndian.Uint32(data[4:8]), data[8:]}, nil
}

// NewCommand : 以命令編號 + proto.Message 類的資料來建立一個 Command 物件
func NewCommand(cmd interface{}, pb proto.Message) *Command {
	// try to marshal given message.
	body, err := proto.Marshal(pb)
	if err != nil {
		Error("Player:Send: invalid data. CMD=%v, ERR=%s", reflect.ValueOf(cmd), err.Error())
		return nil
	}
	// check the command type
	val := reflect.ValueOf(cmd)
	var cmdType uint32
	switch val.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		cmdType = uint32(val.Int())
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		cmdType = uint32(val.Uint())
	default:
		Error("Player:Send: invalid command type. CMD=%v, KIND=%s", val, val.Kind().String())
		return nil
	}
	return &Command{cmdType, uint32(len(body)), body}
}

// Type : Retrieves the command type.
func (c *Command) Type() uint32 {
	return c.cmd
}

// Length : Retrieves the command length -> body
func (c *Command) Length() uint32 {
	return c.length
}

// Data : Retrieves the command body
func (c *Command) Data() []byte {
	return c.body
}

// Bytes : 將 Command 轉化為可以送出的 byte array 資料
func (c *Command) Bytes() (result []byte) {
	result = make([]byte, 8+c.length)
	binary.LittleEndian.PutUint32(result[0:], c.cmd)
	binary.LittleEndian.PutUint32(result[4:], c.length)
	copy(result[8:], c.body)
	return
}
