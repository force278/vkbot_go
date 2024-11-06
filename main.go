package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"vkbot/config"
	"vkbot/database"
	"vkbot/keyboard"
	"vkbot/utils"
)

var keyboards keyboard.Keyboards

// Определяем тип для ключа контекста
type contextKey string

const eventContextKey contextKey = "event"

// LoggingMiddleware - пример middleware для логирования запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event utils.Event
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusBadRequest)
			fmt.Println(err)
			return
		}

		// Десериализация тела в структуру event
		event.GetStruct(body, &event)

		// Логируем информацию о запросе
		log.Printf("Received request: %+v", event)

		// Добавляем event в контекст
		ctx := context.WithValue(r.Context(), eventContextKey, event)
		r = r.WithContext(ctx)

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
	})
}

func main() {
	file, err := os.OpenFile("vkbot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	log.SetOutput(file) // Устанавливаем вывод в файл
	// Загружаем конфигурацию
	if err := config.LoadConfig(&config.AppConfig); err != nil {
		fmt.Println("Ошибка загрузки конфигурации:", err)
		return
	}

	database.Connect()
	defer database.Disconnect()

	if err := keyboards.FromJSON(); err != nil {
		fmt.Println("Ошибка чтения keyboard.json: ", err)
	}

	// Получаем обновления от ВК
	go func() {
		// Создаем новый маршрутизатор
		mux := http.NewServeMux()
		loggedMux := LoggingMiddleware(mux)
		http.HandleFunc("/callback", callbackHandler)
		fmt.Println("Сервер запущен на порту 8080")
		if err := http.ListenAndServe(":8080", loggedMux); err != nil {
			log.Fatal(err)
		}
	}()

	select {} // Бесконечный цикл
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	event, ok := r.Context().Value(eventContextKey).(utils.Event)
	if !ok {
		http.Error(w, "No event found in context", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
	switch event.Type {
	case "message_new":
		/*
			keyboard, _ := keyboards.KeyboardMain.ToJSON()
			funcs.SendMessage(event.Object.Message.FromID, "ABOBA", keyboard)
		*/
		/*
			if len(event.Object.Message.Attachments) > 0 {
				attachment := event.Object.Message.Attachments[0]
				if attachment.Type == "photo" {
					uploadURL := funcs.GetUploadServer()
					if uploadURL != "" {
						funcs.UploadPhoto(uploadURL, *attachment.Photo, event.Object.Message.FromID)
					}
				}
			}
		*/
	case "message_deny":

	case "confirmation":

	default:

	}

}
