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
	"vkbot/funcs"
	"vkbot/keyboard"
	"vkbot/utils"
)

var keyboards keyboard.Keyboards

// Определяем тип для ключа контекста
type contextKey string

const eventContextKey contextKey = "event"
const userContextKey contextKey = "user"

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
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")

		var user utils.User
		var userID uint

		if event.Type == "message_new" {
			userID = event.Object.Message.FromID
			var userExist bool
			user, userExist, err = database.GetUser(userID)
			if err != nil {
				fmt.Printf("Произошла ошибка в Middleware GetUser: %s \n", err)
				return
			}
			if !userExist { // Пользователя нет в бд
				newID, err := database.AddUser(userID)
				if err != nil {
					fmt.Printf("Произошла ошибка в Middleware AddUser: %s \n", err)
					return
				}
				user.UserID = userID // Это все что мы знаем о новом пользователе
				user.ID = newID
			}
		}
		if event.Type == "message_deny" || event.Type == "message_allow" {
			userID = event.Object.Userid
			var userExist bool
			user, userExist, err = database.GetUser(userID)
			if err != nil {
				fmt.Printf("Произошла ошибка в Middleware GetUser: %s \n", err)
				return
			}
			if !userExist { // Пользователя нет в бд
				newID, err := database.AddUser(userID)
				if err != nil {
					fmt.Printf("Произошла ошибка в Middleware AddUser: %s \n", err)
					return
				}
				user.UserID = userID // Это все что мы знаем о новом пользователе
				user.ID = newID
			}
		}
		fmt.Printf("%+v", user)

		// Проверяем запрос
		if user.Ban == 1 {
			return
		}
		// Добавляем event в контекст
		ctx := context.WithValue(r.Context(), eventContextKey, event)
		ctx = context.WithValue(ctx, userContextKey, user)
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
	database.AddStatusColumnIfNotExists()

	if err := keyboards.FromJSON(); err != nil {
		fmt.Println("Ошибка чтения keyboard.json: ", err)
	}

	// Получаем обновления от ВК
	go func() {
		// Создаем новый маршрутизатор
		mux := http.NewServeMux()
		loggedMux := LoggingMiddleware(mux)
		mux.HandleFunc("/callback", callbackHandler)
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

	switch event.Type {
	case "message_new":
		keyboard, err := keyboards.KeyboardMain.ToJSON()
		if err != nil {
			fmt.Printf("Error converting keyboard to JSON: %v", err)
			return
		}
		funcs.SendMessage(event.Object.Message.FromID, "ABOBA", keyboard)
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

	case "message_allow":

	case "confirmation":

	default:

	}

}
