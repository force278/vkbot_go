package funcs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"vkbot/config"
	"vkbot/database"
	"vkbot/keyboard"
	"vkbot/utils"
)

func Handle(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	if user.Admin == 1 {
		switch strings.ToLower(event.Object.Message.Text) {
		case "забанить":
			{
				database.UpdateState(user.UserID, utils.GO_BAN_STATE)
				SendMessage(user.UserID, "Введите UserID пользователя, которого хотите ЗАБАНИТЬ:", "")
				return
			}
		case "разбанить":
			{
				database.UpdateState(user.UserID, utils.GO_UNBAN_STATE)
				SendMessage(user.UserID, "Введите UserID пользователя, которого хотите РАЗБАНИТЬ:", "")
				return
			}
		case "+подписка":
			{
				database.UpdateState(user.UserID, utils.ADD_STATE)
				SendMessage(user.UserID, "Введите UserID пользователя, которому хотите ДОБАВИТЬ подписку:", "")
				return
			}
		case "-подписка":
			{
				database.UpdateState(user.UserID, utils.POP_STATE)
				SendMessage(user.UserID, "Введите UserID пользователя, которому хотите УБРАТЬ подписку:", "")
				return
			}
		case "модер":
			{

			}
		case "-модер":
			{

			}
		case "рассылка123":
			{

			}
		}
	}
	switch user.State {
	case utils.NAME_STATE:
		handleNameState(event, user)
	case utils.PHOTO_STATE:
		handlePhotoState(event, user, keyboards)
	case utils.MENU_STATE:
		handleMenuState(event, user, keyboards)
	case utils.CHANGE_STATE:
		handleChangeState(event, user, keyboards)
	case utils.CHANGE_NAME_STATE:
		handleChangeNameState(event, user)
	case utils.CHANGE_PHOTO_STATE:
		handleChangePhotoState(event, user, keyboards)
	case utils.CHANGE_PHOTO_UPLOAD_STATE:
		handleChangePhotoUploadState(event, user, keyboards)
	case utils.GO_STATE:
		handleGoState(event, user, keyboards)
	case utils.COMPLAINT_STATE:
		handleComplaintState(event, user, keyboards)
	case utils.CHANGE_ADDRESS_STATE:
		handleChangeAddressState(event, user, keyboards)
	case utils.TOP_STATE:
		handleTopState(event, user, keyboards)
	case utils.GO_MESSAGE_STATE:
		handleGoMessageState(event, user, keyboards)
	case utils.GO_MESSAGE_GRADE_STATE:
		handleGoMessageGradeState(event, user, keyboards)
	case utils.GO_BAN_STATE:
		handleBanState(event, user, keyboards)
	case utils.GO_UNBAN_STATE:
		handleUnbanState(event, user, keyboards)
	case utils.ADD_STATE:
		handleAddState(event, user, keyboards)
	case utils.POP_STATE:
		handlePopState(event, user, keyboards)
	default:
		handleDefaultState(user, keyboards)
	}
}

func handleNameState(event utils.Event, user utils.User) {
	message := event.Object.Message.Text
	messageLength := len([]rune(message))
	if messageLength < 2 || messageLength > 20 {
		SendMessage(user.UserID, "Слишком маленькое или слишком длинное имя\n\nВведи другое имя", "")
		return
	}
	pattern := `[,%*&^$£~"#';]`
	re := regexp.MustCompile(pattern)
	if re.MatchString(message) {
		SendMessage(user.UserID, `Ты используешь запрещенные символы: [,%*&^$£~"#';]\nВведи другое имя`, "")
		return
	}
	user.Name = message
	database.UpdateUser(user)
	database.UpdateState(user.UserID, utils.PHOTO_STATE)
	SendMessage(user.UserID, "Теперь отправь фото, которое будут оценивать другие пользователи", "")
}

func handlePhotoState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	if len(event.Object.Message.Attachments) < 1 {
		SendMessage(user.UserID, "Я жду фото", "")
		return
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
		}
	}
}

func handleMenuState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	if event.Object.Message.Payload == "" {
		SendMessage(user.UserID, "Используй клавиатуру", "")
		return
	}
	switch event.Object.Message.Payload {
	case `{"value":"my_profile"}`:
		my_profile(user, keyboards)
	case `{"value":"go_grade"}`:
		handleGoGrade(user, keyboards)
	case `{"value":"my_grades"}`:
		handleMyGrades(user)
	case `{"value":"top"}`:
		handleTop(user, keyboards)
	case `{"value":"about"}`:
		handleAbout(user)
	case `{"value":"menu"}`:
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Меню:", keyboard)
	}

}

func handleGoGrade(user utils.User, keyboards keyboard.Keyboards) {
	rec_user, recExists, err := database.GetRec(user.UserID)
	if err != nil {
		fmt.Printf("Ошибка в MENU_STATE go_grade %s", err)
		return
	}
	if !recExists {
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Больше нет людей для оценки, подождите пока появятся новые пользователи.\n\nМеню:", keyboard)
		return
	}
	message := ""
	if user.Sub == 1 {
		addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", rec_user.UserID, rec_user.Name)
		message = fmt.Sprintf("%s %s", message, addressString)
	}
	var keyboard string
	if user.Admin == 1 {
		keyboard, _ = keyboards.KeyboardGradeModer.ToJSON()
	} else {
		keyboard, _ = keyboards.KeyboardGrade.ToJSON()
	}
	user.RecUser = rec_user.UserID
	user.State = utils.GO_STATE
	database.UpdateUser(user)
	SendPhoto(user.UserID, rec_user.Photo, message, keyboard)
}

func handleMyGrades(user utils.User) {
	grades, err := database.GetGrades(user.UserID)
	if err != nil {
		fmt.Printf("Ошибка в MENU_STATE my_grades: %s", err)
		return
	}
	if len(grades) == 0 {
		SendMessage(user.UserID, "Вас пока никто не оценил, оценивайте чаще и получите оценки.", "")
		return
	}
	for _, grade := range grades {
		if grade.User.Ban == 1 {
			SendMessage(user.UserID, "👮‍♂️Оценка от забаненного пользователя, мы скрыли его.", "")
		}
		message := fmt.Sprintf("🧒Имя оценщика %s\n⭐Оценил вас на %d/10\n", grade.User.Name, grade.Grade)
		if grade.User.Address == 1 || user.Sub == 1 || user.Admin == 1 {
			addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", grade.User.UserID, grade.User.Name)
			message = fmt.Sprintf("%s%s", message, addressString)
		}
		message = fmt.Sprintf("%s%s", message, "👇🏻Фотография оценщика👇🏻")
		SendPhoto(user.UserID, grade.User.Photo, message, "")
	}
}

func handleTop(user utils.User, keyboards keyboard.Keyboards) {
	users, _ := database.Top()
	if len(users) < 1 {
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		database.UpdateState(user.UserID, utils.MENU_STATE)
		SendMessage(user.UserID, "Топ пока не сформирован", keyboard)
		return
	}
	message := fmt.Sprintf("🥇ТОП 1\n\n🍀Имя: %s", users[0].Name)
	if users[0].Address == 1 || user.Admin == 1 || user.Sub == 1 {
		addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", users[0].UserID, users[0].Name)
		message = fmt.Sprintf("%s%s", message, addressString)
	}
	var score float32
	if users[0].People != 0 {
		score = float32(users[0].Score) / float32(users[0].People)
	} else {
		score = 0
	}
	tempMessage := fmt.Sprintf("\n⭐Фото оценили на: %.2f/10\n👥Оценили %d человек", score, users[0].People)
	message = fmt.Sprintf("%s%s", message, tempMessage)
	database.UpdateState(user.UserID, utils.TOP_STATE)
	keyboard, _ := keyboards.KeyboardTop.ToJSON()
	SendPhoto(user.UserID, users[0].Photo, message, keyboard)
}

func handleAbout(user utils.User) {
	message := `👻 Приветствуем тебя в боте, в котором ты сможешь узнать на сколько оценят твою внешность от 1 до 10, и оценить других.\n\n
	💡Если у тебя есть какая-нибудь идея для нашего бота, либо ты нашел баг напиши разработчику @lil_chilllll\n\n
	⚡️Кстати мы делаем приложение бибинто, чекни: bibinto.com `
	SendMessage(user.UserID, message, "")
}

func handleChangeState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch event.Object.Message.Payload {
	case `{"value":"change_name"}`:
		database.UpdateState(user.UserID, utils.CHANGE_NAME_STATE)
		SendMessage(user.UserID, "Введите новое имя:", "")
	case `{"value":"change_photo"}`:
		database.UpdateState(user.UserID, utils.CHANGE_PHOTO_STATE)
		keyboard, _ := keyboards.KeyboardYesNo.ToJSON()
		SendMessage(user.UserID, "Вы точно хотите сменить фото?", keyboard)
	case `{"value":"sub"}`:
		handleSubscription(user, keyboards)
	case `{"value":"buy_check"}`:
		if user.Sub != 1 {
			result, _ := CheckBuySub(user.UserID)
			if result {
				user.Sub = 1
				database.UpdateUser(user)
				keyboard, _ := keyboards.KeyboardProfile.ToJSON()
				SendMessage(user.UserID, "Подписка успешно приобретена!\nТеперь вы будете видеть все ссылки на страницы пользователей.", keyboard)

			}
			SendMessage(user.UserID, "Оплата не найдена", "")
			return
		}
		keyboard, _ := keyboards.KeyboardProfile.ToJSON()
		SendMessage(user.UserID, "У тебя уже есть подписка, ты видишь скрытые ссылки на профили людей.", keyboard)
	case `{"value":"buy"}`:
		buyUrl := "https://yoomoney.ru/quickpay/confirm.xml?receiver=4100117730854038&quickpay-form=shop&targets=Бибинто%20ВК&paymentType=SB&sum=100&label=159236101"
		message := fmt.Sprintf("Перейдите по ссылке, чтобы оплатить подписку\n\nПосле оплаты нажмите кнопку 'Проверить оплату' \n%s\n\nВк может ругаться на подозрительную ссылку, тогда открой ссылку в браузере.", buyUrl)
		SendMessage(user.UserID, message, "")
	case `{"value":"account_link"}`:
		database.UpdateState(user.UserID, utils.CHANGE_ADDRESS_STATE)
		keyboard, _ := keyboards.KeyboardYesNo.ToJSON()
		SendMessage(user.UserID, "Показывать ссылку на вашу страницу другим пользователям?", keyboard)
	case `{"value":"back"}`:
		my_profile(user, keyboards)
	case `{"value":"menu"}`:
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Меню:", keyboard)
	}
}

func handleSubscription(user utils.User, keyboards keyboard.Keyboards) {
	if user.Sub == 1 {
		SendMessage(user.UserID, "У вас уже есть подписка, вы видите скрытые ссылки на профили людей.", "")
		return
	}
	keyboard, _ := keyboards.KeyboardBuySub.ToJSON()
	SendMessage(user.UserID, "Цена подписки 100р (месяц)\n\nПри покупке подписки Вы всегда видите ссылки на страницы людей даже когда оцениваете", keyboard)
}

func handleChangeNameState(event utils.Event, user utils.User) {
	text := event.Object.Message.Text
	messageLength := len([]rune(text))
	if messageLength < 2 || messageLength > 20 {
		SendMessage(user.UserID, "Слишком маленькое или слишком длинное имя\n\nВведи другое имя", "")
		return
	}
	pattern := `[,%*&^$£~"#';]`
	re := regexp.MustCompile(pattern)
	if re.MatchString(text) {
		SendMessage(user.UserID, `Ты используешь запрещенные символы: [,%*&^$£~"#';]\nВведи другое имя`, "")
		return
	}
	user.Name = text
	user.State = utils.CHANGE_STATE
	database.UpdateUser(user)
	SendMessage(user.UserID, "Имя успешно изменено", "")
}

func handleChangePhotoState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch event.Object.Message.Payload {
	case `{"value":"yes"}`:
		database.UpdateState(user.UserID, utils.CHANGE_PHOTO_UPLOAD_STATE)
		SendMessage(user.UserID, "Тогда отправь новое фото", "")
	case `{"value":"no"}`:
		my_profile(user, keyboards)
	case `{"value":"back"}`:
		my_profile(user, keyboards)
	case `{"value":"menu"}`:
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Меню:", keyboard)
	default:
		SendMessage(user.UserID, "Я жду ответа на вопрос...\nЖми на кнопки Да или Нет внизу👇🏻", "")
	}
}

func handleChangePhotoUploadState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	if len(event.Object.Message.Attachments) < 1 {
		SendMessage(user.UserID, "Я жду фото", "")
		return
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
			SendMessage(user.UserID, "Фото успешно изменено.\nМеню:", keyboard)
		}
	}
}

func handleGoState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch event.Object.Message.Payload {
	case `{"value":"menu"}`:
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Меню:", keyboard)
	case `{"value":"grade_report"}`:
		database.UpdateState(user.UserID, utils.COMPLAINT_STATE)
		keyboard, _ := keyboards.KeyboardReportChoose.ToJSON()
		SendMessage(user.UserID, "Введите причину жалобы текстом!\nИли выберите из предложенного", keyboard)
	case `{"value":"grade_message"}`:
		database.UpdateState(user.UserID, utils.GO_MESSAGE_STATE)
		keyboard, _ := keyboards.KeyboardBack.ToJSON()
		SendMessage(user.UserID, "Напиши комментарий к своей оценке:", keyboard)
	case `{"value":"grade_ban"}`:
		handleGradeBan(user, keyboards)
	case `{"value":"grade_1"}`:
		createGrade(1, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_2"}`:
		createGrade(2, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_3"}`:
		createGrade(3, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_4"}`:
		createGrade(4, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_5"}`:
		createGrade(5, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_6"}`:
		createGrade(6, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_7"}`:
		createGrade(7, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_8"}`:
		createGrade(8, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_9"}`:
		createGrade(9, user, "")
		goGrade(user, keyboards, "")
	case `{"value":"grade_10"}`:
		createGrade(10, user, "")
		goGrade(user, keyboards, "")
	}
}

func handleGradeBan(user utils.User, keyboards keyboard.Keyboards) {
	if user.Admin != 1 {
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Ты не админ, чтобы банить", keyboard)
		return
	}
	database.Ban(uint64(user.RecUser))
	message := fmt.Sprintf("Предыдущий пользователь забанен!\nЕго id: %d\n\n📎Ссылка на страницу: @id%d(%s)", user.RecUser, user.RecUser, "Профиль")
	rec_user, recExists, err := database.GetRec(user.UserID)
	if err != nil {
		fmt.Printf("Ошибка в MENU_STATE go_grade %s", err)
		return
	}
	if !recExists {
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Больше нет людей для оценки, подождите пока появятся новые пользователи.\n\nМеню:", keyboard)
		return
	}
	if user.Sub == 1 {
		addressString := fmt.Sprintf("\n\n📎Ссылка на страницу: @id%d(%s)", rec_user.UserID, rec_user.Name)
		message = fmt.Sprintf("%s %s", message, addressString)
	}
	var keyboard string
	if user.Admin == 1 {
		keyboard, _ = keyboards.KeyboardGradeModer.ToJSON()
	} else {
		keyboard, _ = keyboards.KeyboardGrade.ToJSON()
	}
	user.RecUser = rec_user.UserID
	user.State = utils.GO_STATE
	database.UpdateUser(user)
	SendPhoto(user.UserID, rec_user.Photo, message, keyboard)
}

func handleComplaintState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	var adminMessage string
	switch event.Object.Message.Payload {
	case `{"value":"report_18+"}`:
		adminMessage = fmt.Sprintf("Жалоба (18+) от %s|%d на %d", user.Name, user.UserID, user.RecUser)
	case `{"value":"report_younger_14"}`:
		adminMessage = fmt.Sprintf("Жалоба (Младше 14) от %s|%d на %d", user.Name, user.UserID, user.RecUser)
	case `{"value":"spam"}`:
		adminMessage = fmt.Sprintf("Жалоба (Спам) от %s|%d на %d", user.Name, user.UserID, user.RecUser)
	case `{"value":"back"}`:
		goGrade(user, keyboards, "")
	}
	SendMessage(config.AppConfig.ReportAdmin, adminMessage, "")
	goGrade(user, keyboards, "Спасибо за жалобу, мы рассмотрим его в ближайшее время!")
}

func handleChangeAddressState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch event.Object.Message.Payload {
	case `{"value":"yes"}`:
		user.Address = 1
		user.State = utils.CHANGE_STATE
		database.UpdateUser(user)
		keyboard, _ := keyboards.KeyboardProfile.ToJSON()
		SendMessage(user.UserID, "Теперь ваша ссылка ВИДНА другим пользователям.", keyboard)
	case `{"value":"no"}`:
		user.Address = 0
		user.State = utils.CHANGE_STATE
		database.UpdateUser(user)
		keyboard, _ := keyboards.KeyboardProfile.ToJSON()
		SendMessage(user.UserID, "Теперь ваша ссылка НЕ ВИДНА другим пользователям.", keyboard)
	case `{"value":"back"}`:
		my_profile(user, keyboards)
	case `{"value":"menu"}`:
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Меню:", keyboard)
	}
}

func handleTopState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch event.Object.Message.Payload {
	case `{"value":"top_1"}`:
		handleTopPosition(0, user, keyboards)
	case `{"value":"top_2"}`:
		handleTopPosition(1, user, keyboards)
	case `{"value":"top_3"}`:
		handleTopPosition(2, user, keyboards)
	case `{"value":"top_10"}`:
		handleTop10(user, keyboards)
	case `{"value":"my_top_position"}`:
		handleMyTopPosition(user)
	case `{"value":"menu"}`:
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		database.UpdateState(user.UserID, utils.MENU_STATE)
		SendMessage(user.UserID, "Меню", keyboard)
	}
}

func handleTopPosition(index int, user utils.User, keyboards keyboard.Keyboards) {
	users, _ := database.Top()
	if len(users) <= index {
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		database.UpdateState(user.UserID, utils.MENU_STATE)
		SendMessage(user.UserID, "Топ пока не сформирован", keyboard)
		return
	}
	message := fmt.Sprintf("🥇ТОП %d\n\n🍀Имя: %s", index+1, users[index].Name)
	if users[index].Address == 1 || user.Admin == 1 || user.Sub == 1 {
		addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", users[index].UserID, users[index].Name)
		message = fmt.Sprintf("%s%s", message, addressString)
	}
	var score float32
	if users[index].People != 0 {
		score = float32(users[index].Score) / float32(users[index].People)
	} else {
		score = 0
	}
	tempMessage := fmt.Sprintf("\n⭐Фото оценили на: %.2f/10\n👥Оценили %d человек", score, users[index].People)
	message = fmt.Sprintf("%s%s", message, tempMessage)
	keyboard, _ := keyboards.KeyboardTop.ToJSON()
	SendPhoto(user.UserID, users[index].Photo, message, keyboard)
}

func handleTop10(user utils.User, keyboards keyboard.Keyboards) {
	top10, _ := database.Top10()
	var photos string
	for _, photo := range top10 {
		photos = fmt.Sprintf("%s, %s", photos, photo)
	}
	if photos == "" {
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Топ пока не сформирован", keyboard)
		return
	}
	SendPhoto(user.UserID, photos, "", "")
}

func handleMyTopPosition(user utils.User) {
	my_top, _ := database.MyTop(user.UserID)
	message := fmt.Sprintf("Твоя позиция в топе: %d", my_top)
	SendMessage(user.UserID, message, "")
}

func handleGoMessageState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch event.Object.Message.Payload {
	case `{"value":"back"}`:
		rec_user, _, _ := database.GetUser(user.RecUser)
		var message string
		if user.Sub == 1 {
			addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", rec_user.UserID, "Ссылка")
			message = fmt.Sprintf("%s %s", message, addressString)
		}
		var keyboard string
		if user.Admin == 1 {
			keyboard, _ = keyboards.KeyboardGradeModer.ToJSON()
		} else {
			keyboard, _ = keyboards.KeyboardGrade.ToJSON()
		}
		user.RecUser = rec_user.UserID
		user.State = utils.GO_STATE
		database.UpdateUser(user)
		SendPhoto(user.UserID, rec_user.Photo, message, keyboard)
	default:
		handleGoMessageText(event, user, keyboards)
	}
}

func handleGoMessageText(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	if len(event.Object.Message.Text) > 100 {
		SendMessage(user.UserID, "Сообщение слишком длинное.\nНапиши меньше 100 символов.", "")
		return
	}
	if len(event.Object.Message.Text) < 2 {
		SendMessage(user.UserID, "Сообщение слишком короткое.\nНапиши что-нибудь поинтереснее.", "")
		return
	}
	user.State = utils.GO_MESSAGE_GRADE_STATE
	user.RecMess = event.Object.Message.Text
	database.UpdateUser(user)
	var keyboard string
	if user.Admin == 1 {
		keyboard, _ = keyboards.KeyboardGradeModer.ToJSON()
	} else {
		keyboard, _ = keyboards.KeyboardGrade.ToJSON()
	}
	SendMessage(user.UserID, "Теперь поставь оценку:", keyboard)
}

func my_profile(user utils.User, keyboards keyboard.Keyboards) {
	var score float32
	if user.People != 0 {
		score = float32(user.Score) / float32(user.People)
	} else {
		score = 0
	}
	message := fmt.Sprintf("🍀Имя: %s\n\n⭐Ваше фото оценили на: %.2f/10\n👥Вас оценили %d человек", user.Name, score, user.People)
	if user.Address == 1 {
		addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", user.UserID, user.Name)
		message = fmt.Sprintf("%s %s", message, addressString)
	}
	if user.Sub == 1 {
		message = fmt.Sprintf("%s %s", message, "\n⚡Подписка активна⚡")
	}
	database.UpdateState(user.UserID, utils.CHANGE_STATE)
	keyboard, _ := keyboards.KeyboardProfile.ToJSON()
	SendPhoto(user.UserID, user.Photo, message, keyboard)
}

func createGrade(grade int, user utils.User, message string) {
	if message != "" {
		database.AddGrade(user.RecUser, user.UserID, grade, &message)
	} else {
		database.AddGrade(user.RecUser, user.UserID, grade, nil)
	}
}

func goGrade(user utils.User, keyboards keyboard.Keyboards, extraMessage string) {
	rec_user, recExists, err := database.GetRec(user.UserID)
	if err != nil {
		fmt.Printf("Ошибка в goGrade() %s", err)
		return
	}
	if !recExists {
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		message := "Больше нет людей для оценки, подождите пока появятся новые пользователи.\n\nМеню:"
		if extraMessage != "" {
			message = fmt.Sprintf("%s\n\n%s", extraMessage, message)
		}
		SendMessage(user.UserID, message, keyboard)
		return
	}
	var message string
	if extraMessage != "" {
		message = fmt.Sprintf("%s\n\n%s", extraMessage, message)
	}
	if user.Sub == 1 {
		addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", rec_user.UserID, rec_user.Name)
		message = fmt.Sprintf("%s %s", message, addressString)
	}
	var keyboard string
	if user.Admin == 1 {
		keyboard, _ = keyboards.KeyboardGradeModer.ToJSON()
	} else {
		keyboard, _ = keyboards.KeyboardGrade.ToJSON()
	}
	user.RecUser = rec_user.UserID
	user.State = utils.GO_STATE
	database.UpdateUser(user)
	SendPhoto(user.UserID, rec_user.Photo, message, keyboard)
}

func handleGoMessageGradeState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	switch event.Object.Message.Payload {
	case `{"value":"back"}`:
		{
			rec_user, _, _ := database.GetUser(user.RecUser)
			var message string
			if user.Sub == 1 {
				addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id%d(%s)", rec_user.UserID, "Ссылка")
				message = fmt.Sprintf("%s %s", message, addressString)
			}
			var keyboard string
			if user.Admin == 1 {
				keyboard, _ = keyboards.KeyboardGradeModer.ToJSON()
			} else {
				keyboard, _ = keyboards.KeyboardGrade.ToJSON()
			}
			user.RecUser = rec_user.UserID
			user.State = utils.GO_STATE
			database.UpdateUser(user)
			SendPhoto(user.UserID, rec_user.Photo, message, keyboard)
		}
	case `{"value":"grade_report"}`:
		{
			database.UpdateState(user.UserID, utils.COMPLAINT_STATE)
			keyboard, _ := keyboards.KeyboardReportChoose.ToJSON()
			SendMessage(user.UserID, "Введите причину жалобы текстом!\nИли выберите из предложенного", keyboard)
		}
	case `{"value":"grade_ban"}`:
		{
			if user.Admin != 1 {
				database.UpdateState(user.UserID, utils.MENU_STATE)
				keyboard, _ := keyboards.KeyboardMain.ToJSON()
				SendMessage(user.UserID, "Ты не админ, чтобы банить", keyboard)
				return
			}
			database.Ban(uint64(user.RecUser))
			message := fmt.Sprintf("Предыдущий пользователь забанен!\nЕго id: %d\n\n📎Ссылка на страницу: @id%d(%s)", user.RecUser, user.RecUser, "Профиль")
			rec_user, recExists, err := database.GetRec(user.UserID)
			if err != nil {
				fmt.Printf("Ошибка в MENU_STATE go_grade %s", err)
				return
			}
			if !recExists {
				keyboard, _ := keyboards.KeyboardMain.ToJSON()
				SendMessage(user.UserID, "Больше нет людей для оценки, подождите пока появятся новые пользователи.\n\nМеню:", keyboard)
				return
			}
			if user.Sub == 1 {
				addressString := fmt.Sprintf("\n\n📎Ссылка на страницу: @id%d(%s)", rec_user.UserID, rec_user.Name)
				message = fmt.Sprintf("%s %s", message, addressString)
			}
			var keyboard string
			if user.Admin == 1 {
				keyboard, _ = keyboards.KeyboardGradeModer.ToJSON()
			} else {
				keyboard, _ = keyboards.KeyboardGrade.ToJSON()
			}
			user.RecUser = rec_user.UserID
			user.State = utils.GO_STATE
			database.UpdateUser(user)
			SendPhoto(user.UserID, rec_user.Photo, message, keyboard)
		}
	case `{"value":"grade_message"}`:
		{
			SendMessage(user.UserID, "Ты уже написал комментарий к оценке, поставь оценку", "")
		}
	case `{"value":"menu"}`:
		{
			database.UpdateState(user.UserID, utils.MENU_STATE)
			keyboard, _ := keyboards.KeyboardMain.ToJSON()
			SendMessage(user.UserID, "Меню:", keyboard)
		}
	case `{"value":"grade_1"}`:
		{
			createGrade(1, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_2"}`:
		{
			createGrade(2, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_3"}`:
		{
			createGrade(3, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_4"}`:
		{
			createGrade(4, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_5"}`:
		{
			createGrade(5, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_6"}`:
		{
			createGrade(6, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_7"}`:
		{
			createGrade(7, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_8"}`:
		{
			createGrade(8, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_9"}`:
		{
			createGrade(9, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	case `{"value":"grade_10"}`:
		{
			createGrade(10, user, user.RecMess)
			goGrade(user, keyboards, "")
		}
	}
}

func handleDefaultState(user utils.User, keyboards keyboard.Keyboards) {
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
}

func handleBanState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	useridString := event.Object.Message.Text
	userid, err := strconv.ParseUint(useridString, 10, 0) // основание 10, 0 для автоматического выбора размера
	if err != nil {
		SendMessage(user.UserID, "Некорректный ввод id пользователя. Пример: 832787473", "")
		return
	}
	err = database.Ban(userid)
	if err != nil {
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Не получилось забанить пользователя с таким id", keyboard)
		return
	}
	database.UpdateState(user.UserID, utils.MENU_STATE)
	keyboard, _ := keyboards.KeyboardMain.ToJSON()
	SendMessage(user.UserID, "Пользователь успешно забанен", keyboard)

}

func handleUnbanState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	useridString := event.Object.Message.Text
	userid, err := strconv.ParseUint(useridString, 10, 0) // основание 10, 0 для автоматического выбора размера
	if err != nil {
		SendMessage(user.UserID, "Некорректный ввод id пользователя. Пример: 832787473", "")
		return
	}
	err = database.Unban(userid)
	if err != nil {
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Не получилось забанить пользователя с таким id", keyboard)
		return
	}
	database.UpdateState(user.UserID, utils.MENU_STATE)
	keyboard, _ := keyboards.KeyboardMain.ToJSON()
	SendMessage(user.UserID, "Пользователь успешно забанен", keyboard)
}

func handleAddState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	useridString := event.Object.Message.Text
	userid, err := strconv.ParseUint(useridString, 10, 0) // основание 10, 0 для автоматического выбора размера
	if err != nil {
		SendMessage(user.UserID, "Некорректный ввод id пользователя. Пример: 832787473", "")
		return
	}
	err = database.AddSub(userid)
	if err != nil {
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Не получилось добавить подписку пользователю с таким id", keyboard)
		return
	}
	database.UpdateState(user.UserID, utils.MENU_STATE)
	keyboard, _ := keyboards.KeyboardMain.ToJSON()
	SendMessage(user.UserID, "Пользователю успешно добавлена подписка", keyboard)
}

func handlePopState(event utils.Event, user utils.User, keyboards keyboard.Keyboards) {
	useridString := event.Object.Message.Text
	userid, err := strconv.ParseUint(useridString, 10, 0) // основание 10, 0 для автоматического выбора размера
	if err != nil {
		SendMessage(user.UserID, "Некорректный ввод id пользователя. Пример: 832787473", "")
		return
	}
	err = database.PopSub(userid)
	if err != nil {
		database.UpdateState(user.UserID, utils.MENU_STATE)
		keyboard, _ := keyboards.KeyboardMain.ToJSON()
		SendMessage(user.UserID, "Не получилось убрать подписку у пользователя с таким id", keyboard)
		return
	}
	database.UpdateState(user.UserID, utils.MENU_STATE)
	keyboard, _ := keyboards.KeyboardMain.ToJSON()
	SendMessage(user.UserID, "У пользователь успешно убрана подписка", keyboard)
}
