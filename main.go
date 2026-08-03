package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"codeberg.org/Nxus/pulse.nvim/internal"
	"codeberg.org/Nxus/pulse.nvim/internal/discord/discord"
	"codeberg.org/Nxus/pulse.nvim/internal/discord/utils"
)

var (
	re            = regexp.MustCompile(`:\d{1,4}`)
	lastSentAt    time.Time
	pending       []byte
	wsMu          sync.Mutex
	wsConn        *internal.WSConn
	connecting    bool
	serverAddr    string
	validateToken string
	isLocalhost   bool
)

const (
	discordAppID = "1515452982471164065"
)

func handleIdleTimeout(key string, url string) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for range ticker.C {
		if err := internal.Heartbeat(key, url); err != nil {
			logError(err)
		}
	}
}

func cleanUrl(url string) string {
	return re.ReplaceAllString(url, "")
}

func connectWS() error {
	wsMu.Lock()
	if wsConn != nil || connecting {
		wsMu.Unlock()
		return nil
	}
	connecting = true
	wsMu.Unlock()

	defer func() {
		wsMu.Lock()
		connecting = false
		wsMu.Unlock()
	}()

	conn, err := internal.ConnectWS(serverAddr, validateToken, isLocalhost)
	if err != nil {
		return err
	}
	wsMu.Lock()
	if wsConn == nil {
		wsConn = conn
	} else {
		conn.Close()
	}
	wsMu.Unlock()
	return nil
}

func watchWS() {
	for {
		wsMu.Lock()
		conn := wsConn
		wsMu.Unlock()

		if conn == nil {
			if err := connectWS(); err != nil {
				logError(err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if err := conn.Read(); err != nil {
			fmt.Fprintf(os.Stderr, "pulse: ws connection lost: %v\n", err)
			wsMu.Lock()
			if wsConn == conn {
				wsConn = nil
			}
			wsMu.Unlock()
			conn.Close()
			time.Sleep(5 * time.Second)
		}
	}
}

func sendWS(line []byte) {
	for attempt := 0; attempt < 2; attempt++ {
		wsMu.Lock()
		conn := wsConn
		wsMu.Unlock()

		if conn == nil {
			if err := connectWS(); err != nil {
				logError(err)
			}
			continue
		}

		frame := internal.PrepareData(line)
		if err := conn.Write(frame); err != nil {
			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
			wsMu.Lock()
			if wsConn == conn {
				wsConn = nil
			}
			wsMu.Unlock()
			conn.Close()
			continue
		}
		return
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] != "" {
		serverAddr = os.Args[1]
	}
	if len(os.Args) > 2 {
		validateToken = os.Args[2]
	}
	if len(os.Args) > 3 && os.Args[3] == "true" {
		isLocalhost = true
	}

	tim := time.Now().UnixMilli()
	lastSentAt = time.Now()

	if err := connectWS(); err != nil {
		fmt.Fprintf(os.Stderr, "Error connection: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		wsMu.Lock()
		if wsConn != nil {
			wsConn.Close()
		}
		wsMu.Unlock()
	}()

	go watchWS()
	go handleIdleTimeout(validateToken, fmt.Sprintf("https://%s", cleanUrl(serverAddr)))

	flagDiscordIPC := false
	connDiscordIPC := discord.ConnectIPC()
	if connDiscordIPC == nil {
		flagDiscordIPC = true
	}
	if !flagDiscordIPC {
		if err := utils.SendHandshake(connDiscordIPC, discordAppID); err != nil {
			logError(err)
			fmt.Fprintf(os.Stderr, "error send handshake: %v", err)
			flagDiscordIPC = true
		}
		if !flagDiscordIPC {
			var buf = make([]byte, 8)
			if _, err := connDiscordIPC.Read(buf); err != nil {
				logError(err)
				fmt.Fprintf(os.Stderr, "error read connDiscordIPC: %v", err)
				flagDiscordIPC = true
			} else {
				lengthJson := binary.LittleEndian.Uint32(buf[4:8])
				var bufPayload = make([]byte, lengthJson)
				if _, err := io.ReadFull(connDiscordIPC, bufPayload); err != nil {
					logError(err)
					fmt.Fprintf(os.Stderr, "error read payload: %v\n", err)
					flagDiscordIPC = true
				} else if !bytes.Contains(bufPayload, []byte("READY")) {
					fmt.Fprintf(os.Stderr, "error: discord not responde READY")
					flagDiscordIPC = true
				}
			}
		}
	}
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if len(pending) > 0 && !flagDiscordIPC && time.Since(lastSentAt) > 3*time.Second {
				id, err := utils.DynamicUuid()
				if err != nil {
					id = strconv.Itoa(rand.Int())
				}
				if err := discord.GetActivity(connDiscordIPC, pending, id, tim); err != nil {
					fmt.Fprintf(os.Stderr, "pulse: error set activity: %v", err)
				} else {
					lastSentAt = time.Now()
				}
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if bytes.Contains(line, []byte("wsServer")) {
			sendWS(line)
		}
		if bytes.Contains(line, []byte("connect:")) && !flagDiscordIPC {
			pending = bytes.Clone(line)
		}
	}
}
func logError(err error) {
	logPath := "pulse_debug.log"
	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(time.Now().String() + ": " + err.Error() + "\n")
}
