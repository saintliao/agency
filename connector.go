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
	"flag"
	"net/url"
	"reflect"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

//------------------------------------------------------------------------------
//	Constants
//------------------------------------------------------------------------------

const (
	// 最大 i/o 字串數量
	maxMessageSize int64 = 5120
)

//------------------------------------------------------------------------------
//	Structure declare
//------------------------------------------------------------------------------

type (
	// OnCommandMethod : 收命令用的 method
	OnCommandMethod func(cmd *Command)

	// Connector : 連線往 AgencyService的物件
	Connector struct {
		// websocket 連線
		conn *websocket.Conn
		// 寫出資料的 byte slice channel
		message chan []byte
		// 結束旗標
		closeSignal chan struct{}
		// 位置
		address string
		//
		CommandHandler OnCommandMethod
	}
)

//------------------------------------------------------------------------------
//	Public Methods
//------------------------------------------------------------------------------

func NewConnector(address string) *Connector {
	connector := &Connector{
		conn:    nil,
		address: address,
	}
	return connector
}

// Connect : 開始連線
func (c *Connector) Connect() bool {
	if c.conn != nil {
		Error("Connector:Connect: already connect.")
		return false
	}
	var addr = flag.String("addr", c.address, "AgencyService address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		Error("Connector:Connect: cannot connect. ERR=%s", err.Error())
		return false
	}
	c.conn.SetReadLimit(maxMessageSize)
	c.closeSignal = make(chan struct{})
	c.message = make(chan []byte)
	go c.readData()
	go c.writeData()
	return true
}

// Disconnect : 斷線
func (c *Connector) Disconnect() {
	if c.conn == nil {
		Error("Connector:Disconnect: not connect.")
		return
	}
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		Error("Connector:Disconnect: error occur. ERR=%s", err.Error())
		return
	}
	close(c.closeSignal)
	close(c.message)
}

// SendCommand : 送出命令
func (c *Connector) SendCommand(cmd uint32, body []byte) {
	if c.conn == nil {
		Error("Connector:SendCommand: not connect.")
		return
	}
	length := uint32(len(body))
	data := make([]byte, 8+length)
	binary.LittleEndian.PutUint32(data[0:], cmd)
	binary.LittleEndian.PutUint32(data[4:], length)
	copy(data[8:], body)
	c.message <- data
}

// Send : 送出命令給 AgencyServer
// @param	cmd		通訊命令 <- 必須是 uint32 or int32
// @param	pb		通訊協定內容，為 protobuf 中的 Message 型別
func (c *Connector) Send(cmd interface{}, pb proto.Message) error {
	// try to marshal given message.
	body, err := proto.Marshal(pb)
	if err != nil {
		Error("Connector:Send: invalid data. CMD=%v, ERR=%s", reflect.ValueOf(cmd), err.Error())
		return err
	}
	// check the command type
	v := reflect.ValueOf(cmd)
	var cmdType uint32
	switch v.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		cmdType = uint32(v.Int())
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		cmdType = uint32(v.Uint())
	default:
		Error("Connector:Send: invalid command type. CMD=%v, KIND=%s", v, v.Kind().String())
		return errors.New("invalid cmd type")
	}
	// call the old school send method.
	c.SendCommand(cmdType, body)
	return nil
}

// OnCommand : 接收訊息
func (c *Connector) OnCommand(cmd *Command) {
	Info("Connector:OnCommand: CMD=%02d, LEN=%d", cmd.Type(), cmd.Length())
	c.CommandHandler(cmd)
	// switch AgencyToMicro(cmd.Type()) {
	// }
}

//------------------------------------------------------------------------------
//	Private Methods
//------------------------------------------------------------------------------

func (c *Connector) readData() {
	for {
		mt, msg, err := c.conn.ReadMessage()
		if err != nil {
			Error("Connector:readData: error occur. ERR=%s", err.Error())
			c.Disconnect()
			return
		}

		if mt != websocket.BinaryMessage {
			Error("Connector:readData: read with unknown data. MESSAGE_TYPE=%d", mt)
			c.Disconnect()
			return
		}

		var cmd *Command
		if cmd, err = CreateCommand(msg); err != nil {
			Error("Connector:readData: create command failed. ERR=%s", err.Error())
			c.Disconnect()
			return
		}
		c.OnCommand(cmd)
	}
}

func (c *Connector) writeData() {
	for {
		select {
		case <-c.closeSignal:
			return

		case body := <-c.message:
			if err := c.conn.WriteMessage(websocket.BinaryMessage, body); err != nil {
				Error("Connector:writeData: failed. ERR=%s", err.Error())
				c.Disconnect()
			}
		}
	}
}
