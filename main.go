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
			log.Println("Error reading body:", err)
			return
		}
		defer r.Body.Close() // Закрываем тело запроса после чтения

		// Десериализация тела в структуру event
		if err := event.GetStruct(body, &event); err != nil {
			http.Error(w, "Error parsing event", http.StatusBadRequest)
			log.Println("Error parsing event:", err)
			return
		}

		// Логируем информацию о запросе
		log.Printf("Received request: %+v", event)

		var user utils.User
		var userID uint

		// Обработка разных типов событий
		switch event.Type {
		case "message_new":
			userID = event.Object.Message.FromID
		case "message_deny", "message_allow":
			userID = event.Object.Userid
		default:
			http.Error(w, "Unsupported event type", http.StatusBadRequest)
			return
		}

		// Получаем пользователя из базы данных
		var userExist bool
		user, userExist, err = database.GetUser(userID)
		if err != nil {
			log.Printf("Error in Middleware Get:User  %s \n", err)
			return
		}
		if !userExist { // Пользователя нет в БД
			newID, err := database.AddUser(userID)
			if err != nil {
				log.Printf("Error in Middleware Add:User  %s \n", err)
				return
			}
			user.UserID = userID // Это все, что мы знаем о новом пользователе
			user.ID = newID
		}

		log.Printf("User  info: %+v\n", user)

		// Проверяем, забанен ли пользователь
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
		log.Println("Ошибка загрузки конфигурации:", err)
		return
	}

	database.Connect()
	defer database.Disconnect()
	database.AddStateColumnIfNotExists()
	database.AddRecUserColumnIfNotExists()
	database.AddRecMessColumnIfNotExists()

	if err := keyboards.FromJSON(); err != nil {
		log.Println("Ошибка чтения keyboard.json: ", err)
	}

	// Получаем обновления от ВК
	go func() {
		// Создаем новый маршрутизатор
		mux := http.NewServeMux()
		mux.HandleFunc("/callback", callbackHandler)
		loggedMux := LoggingMiddleware(mux)
		log.Println("Сервер запущен на порту 8080")
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
	user, ok := r.Context().Value(userContextKey).(utils.User)
	if !ok {
		http.Error(w, "No user found in context", http.StatusInternalServerError)
		return
	}

	switch event.Type {
	case "message_new":
		funcs.Handle(event, user, keyboards)
		// Обработка вложений, если необходимо
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
		// Обработка события deny
	case "message_allow":
		// Обработка события allow
	case "confirmation":
		// Обработка события confirmation
	default:
		// Обработка неизвестного события
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}
