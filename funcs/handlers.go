package funcs

import (
	"fmt"
	"regexp"
	"vkbot/database"
	"vkbot/keyboard"
	"vkbot/utils"
)

func Handle(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch user.State {
	case utils.NAME_STATE:
		{
			message := event.Object.Message.Text
			messageLength := len([]rune(message))
			if messageLength < 2 || messageLength > 20 {
				SendMessage(user.UserID, "Слишком маленькое или слишком длинное имя\n\nВведи другое имя", "")
				return
			}
			pattern := `[,%*&^$£~"#';]`
			re := regexp.MustCompile(pattern)
			matches := re.MatchString(message)
			if matches {
				SendMessage(user.UserID, `Ты используешь запрещенные символы: [,%*&^$£~"#';]\n\Введи другое имя`, "")
				return
			}
			user.Name = message
			database.UpdateUser(user)
			database.UpdateState(user.UserID, utils.PHOTO_STATE)
			SendMessage(user.UserID, "Теперь отправь фото, которое будут оценивать другие пользователи", "")

		}
	case utils.PHOTO_STATE:
		{
			if len(event.Object.Message.Attachments) < 1 {
				SendMessage(user.UserID, "Я жду фото", "")
			}
			attachment := event.Object.Message.Attachments[0]
			if attachment.Type == "photo" {
				uploadURL := GetUploadServer()
				if uploadURL != "" {
					photo := UploadPhoto(uploadURL, *attachment.Photo, user.UserID)
					user.Photo = photo
					user.State = utils.MENU_STATE
					database.UpdateUser(user)
					keyboard, _ := keyboards.KeyboardMain.ToJSON()
					SendMessage(user.UserID, "Твоя анкета заполнена.\nМеню:", keyboard)
					return
				}
			}

		}
	case utils.MENU_STATE:
		{
			if event.Object.Message.Payload == "" {
				fmt.Print("Поле пустое")
			} else {
				fmt.Print("Поле не пустое")
			}
		}
	case utils.CHANGE_STATE:
		{

		}
	case utils.CHANGE_NAME_STATE:
		{

		}
	case utils.CHANGE_PHOTO_STATE:
		{

		}
	case utils.CHANGE_PHOTO_UPLOAD_STATE:
		{

		}
	case utils.CHECK_STATE:
		{

		}
	case utils.GO_STATE:
		{

		}
	case utils.BAN_STATE:
		{

		}
	case utils.COMPLAINT_STATE:
		{

		}
	case utils.GO_BAN_STATE:
		{

		}
	case utils.CHANGE_ADDRESS_STATE:
		{

		}
	case utils.GO_UNBAN_STATE:
		{

		}
	case utils.TOP_STATE:
		{

		}
	case utils.ADD_STATE:
		{

		}
	case utils.POP_STATE:
		{

		}
	case utils.GO_CHANGE_TOP_STATE:
		{

		}
	case utils.CHANGE_TOP_STATE:
		{

		}
	case utils.GO_MESSAGE_STATE:
		{

		}
	case utils.GO_MESSAGE_GRADE_STATE:
		{

		}
	default:
		{
			// Для начала проверим есть ли имя и фото у пользователя
			if user.Name == "None" || user.Photo == "None" || user.Name == "" || user.Photo == "" {
				// Если у пользователя нет имени или фото, то отправляем заполнять имя
				// Отправляем его в меню
				database.UpdateState(user.UserID, utils.NAME_STATE)
				SendMessage(user.UserID, "Я не смог найти твою анкету.\nДавай заполним ее заново.\n\nНапиши свое имя:", "")
				return
			}
			// Отправляем его в меню
			database.UpdateState(user.UserID, utils.MENU_STATE)
			keyboard, err := keyboards.KeyboardMain.ToJSON()
			if err != nil {
				fmt.Printf("ошибка клавиатуры в handle default %s", err)
				return
			}
			SendMessage(user.UserID, "Меню:", keyboard)
			return
		}
	}
}
