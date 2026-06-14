package discord

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"
)


func SetActivity(conn net.Conn, act Activity, id string) error {
	data := map[string]any{
		"cmd": "SET_ACTIVITY",
		"args": map[string]any{
			"pid":      os.Getgid(),
			"activity": act,
		},
		"nonce": id,
	}
	payload, _ := json.Marshal(data)
	var opcode uint32 = 1
	length := uint32(len(payload))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, opcode)
	binary.Write(buf, binary.LittleEndian, length)
	buf.Write(payload)
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	respHeader := make([]byte, 8)
	if _, err := io.ReadFull(conn, respHeader); err != nil {
		return err
	}
	lengthJson := binary.LittleEndian.Uint32(respHeader[4:8])
	respPayload := make([]byte, lengthJson)
	if _, err := io.ReadFull(conn, respPayload); err != nil {
		return err
	}
	return nil
}
