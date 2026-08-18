package wsserver

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"reflect"

	"github.com/gorilla/websocket"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

func Handle(conn *websocket.Conn, actions map[ActionType]interface{}) error {
	ctx := context.TODO()
	for {
		select {
		case <-lifecycle.Std().C():
			return nil
		default:
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				logs.ErrorContextf(ctx, "read message error: %s", err.Error())
				return err
			}
			// logs.Debugf("receive ws message type: %d", messageType)
			switch messageType {
			case websocket.TextMessage:
			case websocket.BinaryMessage:
			case websocket.PingMessage:
				logs.InfoContextf(ctx, "receive ping message, send pong message")
				conn.WriteMessage(websocket.PongMessage, nil)
				continue
			case websocket.PongMessage:
				logs.InfoContextf(ctx, "receive pong message")
				continue
			case websocket.CloseMessage:
				logs.InfoContextf(ctx, "receive close message, close connection")
				return websocket.ErrCloseSent
			}

			go call(ctx, actions, data)
		}
	}
}

func call(ctx context.Context, actions map[ActionType]interface{}, data []byte) error {
	buf := bytes.NewBuffer(data)
	var evt ActionType
	err := binary.Read(buf, binary.LittleEndian, &evt)
	if err != nil {
		logs.ErrorContextf(ctx, "handleMessage: read message error: %s", err.Error())
		return nil
	}
	// logs.Debugf("handleMessage: receive action type: %d", evt)

	hdr, ok := actions[evt]
	if !ok {
		logs.ErrorContextf(ctx, "handleMessage: action type %d not found", evt)
		return nil
	}

	hdrType := reflect.TypeOf(hdr)
	if hdrType.Kind() != reflect.Func {
		logs.ErrorContextf(ctx, "handleMessage: invalid handler type: %T", hdr)
		return nil
	}

	if hdrType.NumIn() != 1 {
		logs.ErrorContextf(ctx, "invalid handler params number.", hdrType.NumIn())
		return nil
	}
	if hdrType.NumOut() > 1 {
		logs.ErrorContextf(ctx, "invalid handler returns number.")
		return nil
	}

	inVal := reflect.New(hdrType.In(0).Elem())
	in := inVal.Interface()
	if hdrType.In(0).Elem().String() == "bytes.Buffer" {
		inVal.Elem().Set(reflect.ValueOf(buf).Elem())
	} else if err := json.NewDecoder(buf).Decode(in); err != nil {
		logs.ErrorContextf(ctx, "decode request failed, %s", err)
		return nil
	}
	vals := reflect.ValueOf(hdr).Call([]reflect.Value{inVal})
	if len(vals) == 1 {
		retVal := vals[0].Interface()
		if retVal != nil {
			err, ok := retVal.(error)
			if !ok {
				logs.ErrorContextf(ctx, "invalid handler returns type: %T", retVal)
				return nil
			}
			if err != nil {
				logs.ErrorContextf(ctx, "handler return error: %s", err)
				return err
			}
		}
	}

	return nil
}
