package main

import (
	"log"
	"vkbot/config"
	"vkbot/database"
	"vkbot/funcs"
)

func main() {
	// Загружаем конфигурацию
	if err := config.LoadConfigScheduler(&config.AppConfig); err != nil {
		log.Println("Ошибка загрузки конфигурации:", err)
		return
	}
	database.Connect()
	defer database.Disconnect()
	funcs.SendMessageForAll(`
Тебя давно не было и у тебя накопились оценки! Нажимай "Оценивать", чтобы посмотреть их.
	`)
}
