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
		if event.Type == "confirmation" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "29c03f7b")
			return
		}

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
	database.AddAboutColumnIfNotExists()

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
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
		go funcs.Handle(event, user, keyboards)
	case "message_deny":
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
		// Обработка события deny
	case "message_allow":
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
		// Обработка события allow
	default:
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
		// Обработка неизвестного события
	}

}
