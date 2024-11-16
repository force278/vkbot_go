package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
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
		fmt.Println("Ошибка отправки сообщения:", err)
		return
	}
	defer res.Body.Close() // Закрываем тело ответа после обработки

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body) // Читаем тело ответа для диагностики
		fmt.Printf("Ошибка SendMessage: %d, ответ: %s\n", res.StatusCode, string(body))
	}
}

// Отправка фотографии пользователю
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

	//body, _ := io.ReadAll(res.Body) // Читаем тело ответа для диагностики
}

func SendPhotos(userID uint, photos []string, message string, keyboard string) {
	params := url.Values{}
	params.Set("access_token", config.AppConfig.Token)
	params.Set("user_id", fmt.Sprintf("%d", userID))

	// Объединяем все фотографии в одну строку, разделяя их запятыми
	if len(photos) > 0 {
		params.Set("attachment", strings.Join(photos, ","))
	}

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

	//body, _ := io.ReadAll(res.Body) // Читаем тело ответа для диагностики
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

// Сохранение загруженной фотографии
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
		return fmt.Sprintf("photo%d_%d_%s", photo.OwnerID, photo.ID, photo.AccessKey)
	}

	fmt.Println("Ошибка: сохраненная фотография отсутствует в ответе.")
	return ""
}

// Загрузка фотографии на сервер
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
	if _, err = io.Copy(part, fileResp.Body); err != nil {
		fmt.Println("Ошибка копирования файла:", err)
		return ""
	}

	// Закрываем writer, чтобы завершить формирование запроса
	if err = writer.Close(); err != nil {
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

// CheckBuySub отправляет POST-запрос на указанный URL
func CheckBuySub(userid uint) (bool, error) {
	// URL для запроса
	endpoint := "https://yoomoney.ru/api/operation-history"

	// Создаем данные для отправки в формате x-www-form-urlencoded
	data := url.Values{}
	data.Set("type", "deposition")
	data.Set("records", fmt.Sprintf("%d", 30))

	// Создаем новый HTTP-запрос
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", "Bearer "+config.AppConfig.YooMoneyToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close() // Закрываем тело ответа после обработки

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("ошибка: получили статус %s", resp.Status)
	}

	// Обрабатываем ответ
	var response utils.YooMoneyResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}
	// Десериализуем JSON в структуру
	if err := json.Unmarshal(body, &response); err != nil {
		return false, fmt.Errorf("ошибка при десериализации JSON: %v", err)
	}

	for _, operation := range response.Operations {
		//fmt.Printf("Operation ID: %s, Title: %s, Amount: %.2f %s\n", operation.OperationID, operation.Title, operation.Amount, operation.AmountCurrency)
		if operation.Label == string(userid) {
			return true, nil
		}
	}

	return false, nil
}
