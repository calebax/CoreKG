package wsserver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/gorilla/websocket"
)

// ActionType websocket事件类型
type ActionType int16

func SendMessage(conn *websocket.Conn, actionType ActionType, v interface{}) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, actionType)

	err := json.NewEncoder(buf).Encode(v)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
}
