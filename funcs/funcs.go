package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
	"vkbot/config"
	"vkbot/utils"
)

// Отправка сообщения пользователю
func SendMessage(userID uint, message string, keyboard string) {
	params := url.Values{}
	params.Set("access_token", config.AppConfig.Token)
	params.Set("user_id", fmt.Sprintf("%d", userID))
	params.Set("message", message)
	if keyboard != "" {
		params.Set("keyboard", keyboard)
	}
	params.Set("random_id", fmt.Sprintf("%d", time.Now().UnixNano())) // Уникальный ID для каждой отправки сообщения
	params.Set("v", config.AppConfig.ApiVersion)

	res, err := http.PostForm("https://api.vk.com/method/messages.send", params)
	if err != nil {
		fmt.Println("Ошибка отправки сообщения: ", err)
	}
	if res.StatusCode != 200 {
		fmt.Println("Ошибка SendMessage: ", res.StatusCode)
	}
}

func SendPhoto(userID uint, photo string, message string, keyboard string) {
	params := url.Values{}
	params.Set("access_token", config.AppConfig.Token)
	params.Set("user_id", fmt.Sprintf("%d", userID))
	params.Set("attachment", photo)
	params.Set("message", message)
	if keyboard != "" {
		params.Set("keyboard", keyboard)
	}
	params.Set("random_id", fmt.Sprintf("%d", time.Now().UnixNano())) // Уникальный ID для каждой отправки сообщения
	params.Set("v", config.AppConfig.ApiVersion)

	res, err := http.PostForm("https://api.vk.com/method/messages.send", params)
	if err != nil {
		fmt.Println("Ошибка отправки сообщения:", err)
		return
	}
	defer res.Body.Close() // Закрываем тело ответа после обработки

	body, _ := io.ReadAll(res.Body) // Читаем тело ответа для диагностики
	fmt.Println("Ответ от сервера:", string(body))

	// Здесь можно добавить обработку успешного ответа, если это необходимо
}

// Получение URL загрузки для фотографий
func GetUploadServer() string {
	params := url.Values{}
	params.Set("access_token", config.AppConfig.Token)
	params.Set("group_id", config.AppConfig.GroupID)
	params.Set("v", config.AppConfig.ApiVersion)

	resp, err := http.PostForm("https://api.vk.com/method/photos.getMessagesUploadServer", params)
	if err != nil {
		fmt.Println("Ошибка получения URL загрузки:", err)
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Response struct {
			UploadURL string `json:"upload_url"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Ошибка декодирования ответа:", err)
		return ""
	}
	return result.Response.UploadURL
}

func SavePhoto(uploadResult struct {
	Photo  string `json:"photo"`
	Server int    `json:"server"`
	Hash   string `json:"hash"`
}, userID uint) string {
	params := url.Values{}
	params.Set("access_token", config.AppConfig.Token)
	params.Set("v", config.AppConfig.ApiVersion)
	params.Set("photo", uploadResult.Photo)
	params.Set("server", fmt.Sprintf("%d", uploadResult.Server))
	params.Set("hash", uploadResult.Hash)

	// Запрос на сохранение фотографии
	resp, err := http.PostForm("https://api.vk.com/method/photos.saveMessagesPhoto", params)
	if err != nil {
		fmt.Println("Ошибка сохранения фотографии:", err)
		return ""
	}
	defer resp.Body.Close()

	var saveResult struct {
		Response []struct {
			ID        int64  `json:"id"`
			OwnerID   int64  `json:"owner_id"`
			AccessKey string `json:"access_key"`
		} `json:"response"`
		Error struct {
			ErrorCode int    `json:"error_code"`
			ErrorMsg  string `json:"error_msg"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&saveResult); err != nil {
		fmt.Println("Ошибка декодирования ответа:", err)
		return ""
	}

	if saveResult.Error.ErrorCode != 0 {
		fmt.Printf("Ошибка API: код %d, сообщение: %s\n", saveResult.Error.ErrorCode, saveResult.Error.ErrorMsg)
		return ""
	}

	if len(saveResult.Response) > 0 {
		photo := saveResult.Response[0]
		photo_string := fmt.Sprintf("photo%d_%d_%s", photo.OwnerID, photo.ID, photo.AccessKey)
		return photo_string
	} else {
		fmt.Println("Ошибка: сохраненная фотография отсутствует в ответе.")
		return ""
	}
}

func UploadPhoto(uploadURL string, photo utils.Photo, userID uint) string {
	// Получаем файл фотографии
	fileResp, err := http.Get(photo.Sizes[len(photo.Sizes)-1].URL) // Получаем URL фотографии
	if err != nil {
		fmt.Println("Ошибка получения фотографии:", err)
		return ""
	}
	defer fileResp.Body.Close()

	// Создаем буфер для хранения данных
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Добавляем файл в multipart
	part, err := writer.CreateFormFile("photo", "photo.jpg") // Имя файла можно задать любое
	if err != nil {
		fmt.Println("Ошибка создания формы:", err)
		return ""
	}
	// Копируем содержимое файла в часть формы
	_, err = io.Copy(part, fileResp.Body)
	if err != nil {
		fmt.Println("Ошибка копирования файла:", err)
		return ""
	}

	// Закрываем writer, чтобы завершить формирование запроса
	err = writer.Close()
	if err != nil {
		fmt.Println("Ошибка закрытия writer:", err)
		return ""
	}

	// Загружаем фотографию на сервер
	resp, err := http.Post(uploadURL, writer.FormDataContentType(), &buf)
	if err != nil {
		fmt.Println("Ошибка загрузки фотографии:", err)
		return ""
	}
	defer resp.Body.Close()

	// Читаем ответ сервера
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return ""
	}

	// Обрабатываем ответ
	var uploadResult struct {
		Photo  string `json:"photo"`
		Server int    `json:"server"`
		Hash   string `json:"hash"`
	}

	if err := json.Unmarshal(responseBody, &uploadResult); err != nil {
		fmt.Println("Ошибка декодирования ответа:", err)
		return ""
	}

	// Проверяем, что данные корректны
	if uploadResult.Photo == "[]" || uploadResult.Server == 0 || uploadResult.Hash == "" {
		fmt.Printf("Ошибка загрузки фотографии: %+v\n", uploadResult)
		return ""
	}
	return SavePhoto(uploadResult, userID)
}

func CheckBuySub() {

}
