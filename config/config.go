package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Token       string `json:"token"`
	GroupID     string `json:"groupID"`
	ApiVersion  string `json:"apiVersion"`
	DbName      string `json:"dbName"`
	DbUser      string `json:"dbUser"`
	DbPassword  string `json:"dbPassword"`
	DbHost      string `json:"dbHost"`
	DbPort      uint16 `json:"dbPort"`
	reportAdmin uint   `json:"reportAdmin"`
}

var AppConfig Config

// Функция для загрузки конфигурации из файла
func LoadConfig(AppConfig *Config) error {
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&AppConfig)
}
