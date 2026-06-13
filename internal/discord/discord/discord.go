package discord

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"codeberg.org/Nxus/pulse.nvim/internal/discord/utils"
	"github.com/hugolgst/rich-go/client"
)

type Activity struct {
	Details    string    `json:"details"`
	State      string    `json:"state"`
	LargeImage string    `json:"large_image"`
	Timestams  *TimeStap `json:"timesptamps,omitempty"`
}

type TimeStap struct {
	Start int64 `json:"start,omitempty"`
}

func UpdateDiscordPresence(discordAppID *string, t *time.Time, filename *string, gitRepo *string, isRedacted bool) error {
	if discordAppID == nil || len(*discordAppID) == 0 {
		panic("Discord App ID is required")
	}

	fileExtension := ""
	if filename != nil {
		fileExtensionParts := strings.Split(*filename, ".")
		if len(fileExtensionParts) > 1 {
			fileExtension = fileExtensionParts[len(fileExtensionParts)-1]
		}
	}
	extensionImage := utils.GetLanguageUrl(fmt.Sprintf(".%s", fileExtension), isRedacted)
	err := client.Login(*discordAppID)
	if err != nil {
		return err
	}
	activity := client.Activity{
		LargeImage: extensionImage,
		SmallImage: "https://www.iconarchive.com/show/papirus-apps-icons-by-papirus-team/nvim-icon.html",
		SmallText:  "NeoVim the best editor ever",
		Details:    "Idle",
		Timestamps: &client.Timestamps{
			Start: t,
		},
	}
	if gitRepo != nil || !isRedacted {
		activity.Details = fmt.Sprintf("Workspace: %s", *gitRepo)
	} else if isRedacted {
		activity.Details = ""
	}
	if filename != nil {
		if isRedacted {
			activity.State = "[REDACTED]"
		} else {
			activity.State = fmt.Sprintf("Editing: %s", *filename)
		}
	}
	err = client.SetActivity(activity)
	if err != nil {
		return err
	}
	return nil
}

func SetActivity(conn net.Conn, activity string, id string, startTime int64) error {
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
	return err
}
