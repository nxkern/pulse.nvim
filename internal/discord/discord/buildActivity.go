package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"codeberg.org/Nxus/pulse.nvim/internal/discord/utils"
)

type DataRecived struct {
	Workspace string `json:"workspace"`
	File      string `json:"file"`
	Filetype  string `json:"filetype,omitempty"`
}
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

var (
	fistSent uint8 = 0
	file           = ""
)

func GetActivity(conn net.Conn, data []byte, id string, timer int64) error {
	if conn == nil {
		return fmt.Errorf("pulse: discord not connected")
	}
	var (
		dataR    DataRecived
		activity Activity
	)
	jsonPart := bytes.TrimPrefix(data, []byte("connect:"))

	if err := json.Unmarshal(jsonPart, &dataR); err != nil {
		return fmt.Errorf("pulse: error unmarshal data: %v", err)
	}
	if file == dataR.File && fistSent > 0 {
		return nil
	}

	if dataR.File == "" && fistSent > 0 {
		return nil
	}
	fistSent = 1
	extensions := utils.GetKnowLangs()
	if dataR.File == "" {
		dataR.File = "homepage"
	}
	lang := utils.DetectingLang(dataR.File, extensions)
	activity.Timestams = &TimeStap{
		Start: timer,
	}
	activity.Details = func() string {
		return fmt.Sprintf(func() string {
			if lang == "managerFile" || lang == "LICENSE" {
				return "View: %s"
			}
			if lang == "idle" || lang == "homepage" {
				return "%s"
			}
			if lang == "term" {
				return "Terminal: %s"
			}
			return "Editing: %s"
		}(), dataR.File)
	}()
	activity.State = fmt.Sprintf("Workspace: %s", dataR.Workspace)
	activity.Assets = &Assets{
		LargeImage: fmt.Sprintf("https://codeberg.org/Nxus/pulse.nvim/raw/branch/main/assets/icons/%s.png", func() string {
			if lang == "" {
				return "text"
			}
			return lang
		}()),
		LargeImageText: lang,
		SmallImage:     "https://icons.iconarchive.com/icons/papirus-team/papirus-apps/256/nvim-icon.png",
		SmallImageText: "The best editor Neovim",
	}
	if err := SetActivity(conn, activity, id); err != nil {
		return err
	}
	file = dataR.File
	return nil
}
