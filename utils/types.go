package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// Определяем структуры для десериализации JSON
type Event struct {
	Type   string      `json:"type"`
	V      string      `json:"v"`
	Object EventObject `json:"object"`
}

type EventObject struct {
	Message    Message    `json:"message,omitempty"`
	Userid     uint       `json:"user_id,omitempty"`
	ClientInfo ClientInfo `json:"client_info,omitempty"`
}

type Message struct {
	FromID      uint         `json:"from_id"`
	ID          uint         `json:"id"`
	IsHidden    bool         `json:"is_hidden"`
	Attachments []Attachment `json:"attachments"`
	Text        string       `json:"text"`
	Payload     string       `json:"payload,omitempty"`
}

type Attachment struct {
	Type         string        `json:"type"`
	Photo        *Photo        `json:"photo,omitempty"`
	AudioMessage *AudioMessage `json:"audio_message,omitempty"`
}

type Photo struct {
	AlbumID   int64       `json:"album_id"`
	ID        uint        `json:"id"`
	OwnerID   int64       `json:"owner_id"`
	AccessKey string      `json:"access_key"`
	Sizes     []PhotoSize `json:"sizes"`
	Text      string      `json:"text"`
	OrigPhoto OrigPhoto   `json:"orig_photo"`
}

type PhotoSize struct {
	Height int    `json:"height"`
	Type   string `json:"type"`
	Width  int    `json:"width"`
	URL    string `json:"url"`
}

type OrigPhoto struct {
	Height int    `json:"height"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

type AudioMessage struct {
	Duration  int    `json:"duration"`
	ID        uint   `json:"id"`
	LinkMP3   string `json:"link_mp3"`
	LinkOGG   string `json:"link_ogg"`
	OwnerID   uint   `json:"owner_id"`
	AccessKey string `json:"access_key"`
	Waveform  []int  `json:"waveform"`
}

type ClientInfo struct {
	ButtonActions  []string `json:"button_actions"`
	Keyboard       bool     `json:"keyboard"`
	InlineKeyboard bool     `json:"inline_keyboard"`
	LangID         int      `json:"lang_id"`
}

func (Event) GetStruct(data []byte, event *Event) error {
	// Проверяем, что данные не пустые
	if len(data) == 0 {
		return fmt.Errorf("empty data provided")
	}

	err := json.Unmarshal(data, event)
	if err != nil {
		return fmt.Errorf("error deserializing JSON: %w", err)
	}
	return nil
}

// User представляет структуру для хранения информации о пользователе
type User struct {
	ID          uint
	UserID      uint
	Name        string
	Photo       string
	Score       int
	People      int
	Active      int
	Ban         int
	Address     int
	Admin       int
	Sub         int
	LastMessage time.Time
	State       int
	RecUser     uint
}

type Grade struct {
	ID       uint
	UserID   uint
	ValuerID uint
	Grade    int
	Message  string
	User     User
	Valuer   User
}

type History struct {
	ID       uint
	UserID   uint
	ValuerID int
}

type Stack struct {
	ID     uint
	UserID uint
}
