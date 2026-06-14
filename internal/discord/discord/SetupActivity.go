package discord

import (
	"fmt"
	"net"
	"os"

	"codeberg.org/Nxus/pulse.nvim/internal/discord/utils"
)

var (
	DiscordAppID = ""
)

const (
	TIMEOUT_DURATION = 30
	REDACT_COMMAND   = "redact"
	UNREDACT_COMMAND = "unredact"
)

// func HandleInitialCLientConnection(message string, startTime time.Time) bool {
// 	if strings.Contains(message, "connect") {
// 		onConnectArguments := strings.Split(message, ":")
// 		DiscordAppID = onConnectArguments[1]
// 		fmt.Fprintf(os.Stdout, "Message: %s\n", message)
// 		UpdateDiscordPresence(&DiscordAppID, &startTime, nil, nil, false)
// 		return true
// 	}
// 	return false
// }

func ConnectIPC() net.Conn  {
	var (
		runtimeDir = utils.GetRuntimeDir()
		conn net.Conn
		err error
	)
	for n := range 10 {
		conn, err = net.Dial("unix", fmt.Sprintf("%s/discord-ipc-%d", runtimeDir,n))
		if err != nil {
			continue
		}
		break
	}
	if conn == nil {
		fmt.Fprint(os.Stderr, "error conn nil")
	}
	return conn
}
