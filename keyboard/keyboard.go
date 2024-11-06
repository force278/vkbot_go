package keyboard

import (
	"encoding/json"
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

// ToJSON преобразует клавиатуру в JSON формат
func (k *Keyboard) ToJSON() (string, error) {
	jsonData, err := json.Marshal(k)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
