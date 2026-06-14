package discord

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Activity struct {
	Details   string    `json:"details"`
	State     string    `json:"state"`
	Assets    *Assets   `json:"assets,omitempty"`
	Timestams *TimeStap `json:"timestamps,omitempty"`
}

type Assets struct {
	LargeImage     string `json:"large_image"`
	LargeImageText string `json:"large_text,omitempty"`
	SmallImage     string `json:"small_image"`
	SmallImageText string `json:"small_text,omitempty"`
}

type TimeStap struct {
	Start int64 `json:"start,omitempty"`
}

func SetActivity(conn net.Conn, activity string, id string, startTime int64) error {
	if conn == nil {
		return fmt.Errorf("discord not connected")
	}
	var printOnc sync.Once
	jsonPart := strings.TrimPrefix(activity, "connect:")
	var act Activity
	if err := json.Unmarshal([]byte(jsonPart), &act); err != nil {
		return err
	}
	act.Timestams = &TimeStap{
		Start: startTime,
	}
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
	printOnc.Do(func() {
		fmt.Println("Discord Connect")
	})
	return nil
}
