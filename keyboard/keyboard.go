package keyboard

import (
	"encoding/json"
	"fmt"
	"os"
)

// Keyboards содержит все клавиатуры
type Keyboards struct {
	KeyboardMain         Keyboard `json:"keyboard_main"`
	KeyboardProfile      Keyboard `json:"keyboard_profile"`
	KeyboardYesNo        Keyboard `json:"keyboard_yes_no"`
	KeyboardTop          Keyboard `json:"keyboard_top"`
	KeyboardGrade        Keyboard `json:"keyboard_grade"`
	KeyboardGradeModer   Keyboard `json:"keyboard_grade_moder"`
	KeyboardReportChoose Keyboard `json:"keyboard_report_choose"`
	KeyboardBack         Keyboard `json:"keyboard_back"`
	KeyboardBuySub       Keyboard `json:"keyboard_buy_sub"`
	KeyboardChangeAbout  Keyboard `json:"keyboard_change_about"`
}

// FromJSON преобразует JSON в структуру Keyboards
func (kb *Keyboards) FromJSON() error {
	file, err := os.Open("./keyboard/keyboards.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Decode(&kb)
	return nil
}

// ToJSON преобразует структуру Keyboards в JSON
func (k *Keyboards) ToJSON() (string, error) {
	jsonData, err := json.Marshal(k)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

type Keyboard struct {
	Buttons [][]Button `json:"buttons"`
	OneTime bool       `json:"one_time,omitempty"`
	Inline  bool       `json:"inline,omitempty"`
}

// Button представляет собой кнопку на клавиатуре
type Button struct {
	Action Action `json:"action"`
	Color  string `json:"color,omitempty"`
}

// Action представляет собой действие кнопки
type Action struct {
	Type    string `json:"type"`
	Label   string `json:"label"`
	Payload string `json:"payload,omitempty"`
	Link    string `json:"link,omitempty"`
}

type PayloadData struct {
	UserID int    `json:"user_id,omitempty"`
	Value  string `json:"value"`
}

// NewKeyboard создает новую клавиатуру
func NewKeyboard(oneTime bool, inline bool) *Keyboard {
	return &Keyboard{
		Buttons: make([][]Button, 0),
		OneTime: oneTime,
		Inline:  inline,
	}
}

// AddButton добавляет кнопку на клавиатуру
func (k *Keyboard) AddButton(label string, buttonType string, color string) {
	button := Button{
		Action: Action{
			Type:  buttonType,
			Label: label,
		},
		Color: color,
	}
	k.Buttons = append(k.Buttons, []Button{button})
}

// AddRow добавляет ряд кнопок на клавиатуру
func (k *Keyboard) AddRow(buttons ...Button) {
	k.Buttons = append(k.Buttons, buttons)
}

// AddToPayload добавляет ключ-значение в payload кнопки
func (b *Button) AddToPayload(key string, value interface{}) error {
	// Если payload пустой, создаем новый мап
	if b.Action.Payload == "" {
		b.Action.Payload = "{}"
	}

	// Создаем временную структуру для работы с JSON
	var payloadMap map[string]interface{}
	err := json.Unmarshal([]byte(b.Action.Payload), &payloadMap)
	if err != nil {
		return err
	}

	// Добавляем новый ключ-значение в мапу
	payloadMap[key] = value

	// Сериализуем обратно в JSON
	updatedPayload, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	// Обновляем поле Payload
	b.Action.Payload = string(updatedPayload)
	return nil
}

// ReadFromPayload считывает данные из payload в структуру PayloadData
func (b *Button) ReadFromPayload() (PayloadData, error) {
	var data PayloadData

	if b.Action.Payload == "" {
		return data, fmt.Errorf("payload is empty")
	}

	err := json.Unmarshal([]byte(b.Action.Payload), &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// ToJSON преобразует клавиатуру в JSON формат
func (k *Keyboard) ToJSON() (string, error) {
	jsonData, err := json.Marshal(k)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
