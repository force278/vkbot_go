package funcs

import (
	"fmt"
	"regexp"
	"vkbot/config"
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
				SendMessage(user.UserID, "Используй клавиатуру", "")
			} else {
				switch event.Object.Message.Payload {
				case `{"value":"my_profile"}`:
					{
						my_profile(user, keyboards)
					}
				case `{"value":"go_grade"}`:
					{
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
						var message string
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
				case `{"value":"my_grades"}`:
					{
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
				case `{"value":"top"}`:
					{
						users, _ := database.Top()
						if len(users) < 1 {
							keyboard, _ := keyboards.KeyboardMain.ToJSON()
							database.UpdateState(user.UserID, utils.MENU_STATE)
							SendMessage(user.UserID, "Топ пока не сформирован", keyboard)
							return
						}
						message := fmt.Sprintf("🥇ТОП 1\n\n🍀Имя: %s", users[0].Name)
						if users[0].Address == 1 || user.Admin == 1 || user.Sub == 1 {
							addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id{%d}({%s})", users[0].UserID, users[0].Name)
							message = fmt.Sprintf("%s%s", message, addressString)
						}
						var score float32
						if users[0].People != 0 {
							score = float32(users[0].Score) / float32(users[0].People)
						} else {
							score = 0
						}
						tempMessage := fmt.Sprintf("\n⭐Фото оценили на: {%.2f}/10\n👥Оценили {%d} человек", score, users[0].People)
						message = fmt.Sprintf("%s%s", message, tempMessage)
						keyboard, _ := keyboards.KeyboardTop.ToJSON()
						SendPhoto(user.UserID, users[0].Photo, message, keyboard)
					}
				case `{"value":"about"}`:
					{
						message := `👻 Приветствуем тебя в боте, в котором ты сможешь узнать на сколько оценят твою внешность от 1 до 10, и оценить других.\n\n
        💡Если у тебя есть какая-нибудь идея для нашего бота, либо ты нашел баг напиши разработчику @lil_chilllll\n\n
        ⚡️Кстати мы делаем приложение бибинто, чекни: bibinto.com `
						SendMessage(user.UserID, message, "")
					}
				}

			}
		}
	case utils.CHANGE_STATE:
		{
			switch event.Object.Message.Payload {
			case `{"value":"change_name"}`:
				{
					database.UpdateState(user.UserID, utils.CHANGE_NAME_STATE)
					SendMessage(user.UserID, "Введите новое имя:", "")
				}
			case `{"value":"change_photo"}`:
				{
					database.UpdateState(user.UserID, utils.CHANGE_PHOTO_STATE)
					SendMessage(user.UserID, "Вы точно хотите сменить фото?", "")
				}
			case `{"value":"sub"}`:
				{
					if user.Sub == 1 {
						SendMessage(user.UserID, "У вас уже есть подписка, вы видите скрытые ссылки на профили людей.", "")
						return
					}
					keyboard, _ := keyboards.KeyboardBuySub.ToJSON()
					SendMessage(user.UserID, "Цена подписки 100р (месяц)\n\nПри покупке подписки Вы всегда видете ссылки на страницы людей даже когда оцениваете", keyboard)
				}
			case `{"value":"buy_check"}`:
				{
					CheckBuySub()
					SendMessage(user.UserID, "Заглушка", "")
				}
			case `{"value":"buy"}`:
				{
					var buyUrl string
					message := fmt.Sprintf("Перейдите по ссылке, чтобы оплатить подписку\n\nПосле оплаты нажмите кнопку 'Проверить оплату' \n%s", buyUrl)
					SendMessage(user.UserID, message, "")
				}
			case `{"value":"account_link"}`:
				{
					database.UpdateState(user.UserID, utils.CHANGE_ADDRESS_STATE)
					keyboard, _ := keyboards.KeyboardYesNo.ToJSON()
					SendMessage(user.UserID, "Показывать ссылку на вашу страницу другим пользователям?", keyboard)
				}
			case `{"value":"back"}`:
				{
					my_profile(user, keyboards)
				}
			case `{"value":"menu"}`:
				{
					database.UpdateState(user.UserID, utils.MENU_STATE)
					keyboard, _ := keyboards.KeyboardMain.ToJSON()
					SendMessage(user.UserID, "Меню:", keyboard)
				}
			}
		}
	case utils.CHANGE_NAME_STATE:
		{
			text := event.Object.Message.Text
			messageLength := len([]rune(text))
			if messageLength < 2 || messageLength > 20 {
				SendMessage(user.UserID, "Слишком маленькое или слишком длинное имя\n\nВведи другое имя", "")
				return
			}
			pattern := `[,%*&^$£~"#';]`
			re := regexp.MustCompile(pattern)
			matches := re.MatchString(text)
			if matches {
				SendMessage(user.UserID, `Ты используешь запрещенные символы: [,%*&^$£~"#';]\n\Введи другое имя`, "")
				return
			}
			user.Name = text
			user.State = utils.CHANGE_STATE
			database.UpdateUser(user)
			SendMessage(user.UserID, "Имя успешно изменено", "")
		}
	case utils.CHANGE_PHOTO_STATE:
		{
			switch event.Object.Message.Payload {
			case `{"value":"yes"}`:
				{
					database.UpdateState(user.UserID, utils.CHANGE_PHOTO_UPLOAD_STATE)
					SendMessage(user.UserID, "Тогда отправь новое фото", "")
				}
			case `{"value":"no"}`:
				{
					my_profile(user, keyboards)
				}
			default:
				{
					SendMessage(user.UserID, "Я жду ответа на вопрос...\nЖми на кнопки Да или Нет внизу👇🏻", "")
				}
			}
		}
	case utils.CHANGE_PHOTO_UPLOAD_STATE:
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
					SendMessage(user.UserID, "Фото успешно изменено.\nМеню:", keyboard)
					return
				}
			}
		}
	case utils.GO_STATE:
		{
			switch event.Object.Message.Payload {
			case `{"value":"menu"}`:
				{
					database.UpdateState(user.UserID, utils.MENU_STATE)
					keyboard, _ := keyboards.KeyboardMain.ToJSON()
					SendMessage(user.UserID, "Меню:", keyboard)
				}
			case `{"value":"grade_report"}`:
				{
					database.UpdateState(user.UserID, utils.COMPLAINT_STATE)
					keyboard, _ := keyboards.KeyboardReportChoose.ToJSON()
					SendMessage(user.UserID, "Введите причину жалобы текстом!\nИли выберите из предложенного", keyboard)
				}
			case `{"value":"grade_message"}`:
				{
					database.UpdateState(user.UserID, utils.GO_MESSAGE_STATE)
					keyboard, _ := keyboards.KeyboardBack.ToJSON()
					SendMessage(user.UserID, "Напиши комментарий к своей оценке:", keyboard)
				}
			case `{"value":"grade_ban"}`:
				{
					if user.Admin != 1 {
						database.UpdateState(user.UserID, utils.MENU_STATE)
						keyboard, _ := keyboards.KeyboardMain.ToJSON()
						SendMessage(user.UserID, "Ты не админ, чтобы банить", keyboard)
						return
					}
					database.Ban(user.RecUser)
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
			case `{"value":"grade_1"}`:
				{
					createGrade(1, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_2"}`:
				{
					createGrade(2, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_3"}`:
				{
					createGrade(3, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_4"}`:
				{
					createGrade(4, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_5"}`:
				{
					createGrade(5, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_6"}`:
				{
					createGrade(6, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_7"}`:
				{
					createGrade(7, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_8"}`:
				{
					createGrade(8, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_9"}`:
				{
					createGrade(9, user, "")
					goGrade(user, keyboards, "")
				}
			case `{"value":"grade_10"}`:
				{
					createGrade(10, user, "")
					goGrade(user, keyboards, "")
				}
			}
		}
	case utils.COMPLAINT_STATE:
		{
			var adminMessage string
			switch event.Object.Message.Payload {
			case `{"value":"report_18+"}`:
				{
					adminMessage = fmt.Sprintf("Жалоба (18+) от %s|%d на user_id", user.Name, user.UserID, user.RecUser)
				}
			case `{"value":"report_younger_14"}`:
				{
					adminMessage = fmt.Sprintf("Жалоба (Младше 14) от %s|%d на user_id", user.Name, user.UserID, user.RecUser)
				}
			case `{"value":"spam"}`:
				{
					adminMessage = fmt.Sprintf("Жалоба (Спам) от %s|%d на user_id", user.Name, user.UserID, user.RecUser)
				}
			case `{"value":"back"}`:
				{
					goGrade(user, keyboards, "")
				}
				SendMessage(config.AppConfig.reportAdmin, adminMessage, "")
				goGrade(user, keyboards, "Спасибо за жалобу, мы рассмотрим его в ближайшее время!")
			}

		}
	case utils.GO_BAN_STATE:
		{

		}
	case utils.CHANGE_ADDRESS_STATE:
		{
			switch event.Object.Message.Payload {
			case `{"value":"yes"}`:
				{
					user.Address = 1
					user.State = utils.CHANGE_STATE
					database.UpdateUser(user)
					keyboard, _ := keyboards.KeyboardProfile.ToJSON()
					SendMessage(user.UserID, "Теперь ваша ссылка ВИДНА другим пользователям.", keyboard)
				}
			case `{"value":"no"}`:
				{
					user.Address = 0
					user.State = utils.CHANGE_STATE
					database.UpdateUser(user)
					keyboard, _ := keyboards.KeyboardProfile.ToJSON()
					SendMessage(user.UserID, "Теперь ваша ссылка НЕ ВИДНА другим пользователям.", keyboard)
				}
			case `{"value":"back"}`:
				{
					my_profile(user, keyboards)
				}
			case `{"value":"menu"}`:
				{
					database.UpdateState(user.UserID, utils.MENU_STATE)
					keyboard, _ := keyboards.KeyboardMain.ToJSON()
					SendMessage(user.UserID, "Меню:", keyboard)
				}
			}
		}
	case utils.GO_UNBAN_STATE:
		{

		}
	case utils.TOP_STATE:
		{
			switch event.Object.Message.Payload {
			case `{"value":"top_1"}`:
				{
					users, _ := database.Top()
					if len(users) < 1 {
						keyboard, _ := keyboards.KeyboardMain.ToJSON()
						database.UpdateState(user.UserID, utils.MENU_STATE)
						SendMessage(user.UserID, "Топ пока не сформирован", keyboard)
						return
					}
					message := fmt.Sprintf("🥇ТОП 1\n\n🍀Имя: %s", users[0].Name)
					if users[0].Address == 1 || user.Admin == 1 || user.Sub == 1 {
						addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id{%d}({%s})", users[0].UserID, users[0].Name)
						message = fmt.Sprintf("%s%s", message, addressString)
					}
					var score float32
					if users[0].People != 0 {
						score = float32(users[0].Score) / float32(users[0].People)
					} else {
						score = 0
					}
					tempMessage := fmt.Sprintf("\n⭐Фото оценили на: {%.2f}/10\n👥Оценили {%d} человек", score, users[0].People)
					message = fmt.Sprintf("%s%s", message, tempMessage)
					keyboard, _ := keyboards.KeyboardTop.ToJSON()
					SendPhoto(user.UserID, users[0].Photo, message, keyboard)
				}
			case `{"value":"top_2"}`:
				{
					users, _ := database.Top()
					if len(users) < 2 {
						keyboard, _ := keyboards.KeyboardMain.ToJSON()
						database.UpdateState(user.UserID, utils.MENU_STATE)
						SendMessage(user.UserID, "Топ пока не сформирован", keyboard)
						return
					}
					message := fmt.Sprintf("🥇ТОП 1\n\n🍀Имя: %s", users[1].Name)
					if users[1].Address == 1 || user.Admin == 1 || user.Sub == 1 {
						addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id{%d}({%s})", users[1].UserID, users[1].Name)
						message = fmt.Sprintf("%s%s", message, addressString)
					}
					var score float32
					if users[1].People != 0 {
						score = float32(users[1].Score) / float32(users[1].People)
					} else {
						score = 0
					}
					tempMessage := fmt.Sprintf("\n⭐Фото оценили на: {%.2f}/10\n👥Оценили {%d} человек", score, users[1].People)
					message = fmt.Sprintf("%s%s", message, tempMessage)
					keyboard, _ := keyboards.KeyboardTop.ToJSON()
					SendPhoto(user.UserID, users[1].Photo, message, keyboard)
				}
			case `{"value":"top_3"}`:
				{
					users, _ := database.Top()
					if len(users) < 3 {
						keyboard, _ := keyboards.KeyboardMain.ToJSON()
						database.UpdateState(user.UserID, utils.MENU_STATE)
						SendMessage(user.UserID, "Топ пока не сформирован", keyboard)
						return
					}
					message := fmt.Sprintf("🥇ТОП 1\n\n🍀Имя: %s", users[2].Name)
					if users[2].Address == 1 || user.Admin == 1 || user.Sub == 1 {
						addressString := fmt.Sprintf("\n📎Ссылка на страницу: @id{%d}({%s})", users[2].UserID, users[2].Name)
						message = fmt.Sprintf("%s%s", message, addressString)
					}
					var score float32
					if users[2].People != 0 {
						score = float32(users[2].Score) / float32(users[2].People)
					} else {
						score = 0
					}
					tempMessage := fmt.Sprintf("\n⭐Фото оценили на: {%.2f}/10\n👥Оценили {%d} человек", score, users[2].People)
					message = fmt.Sprintf("%s%s", message, tempMessage)
					keyboard, _ := keyboards.KeyboardTop.ToJSON()
					SendPhoto(user.UserID, users[0].Photo, message, keyboard)
				}
			case `{"value":"top_10"}`:
				{
					top10, _ := database.Top10()
					var photos string
					for photo := range top10 {
						photos = fmt.Sprintf("%s, %s", photos, photo)
					}
					SendPhoto(user.UserID, photos, "", "")
				}
			case `{"value":"my_top_position"}`:
				{
					my_top, _ := database.MyTop(user.UserID)
					message := fmt.Sprintf("Твоя позиция в топе: %d", my_top)
					SendMessage(user.UserID, message, "")
				}
			case `{"value":"menu"}`:
				{
					keyboard, _ := keyboards.KeyboardMain.ToJSON()
					database.UpdateState(user.UserID, utils.MENU_STATE)
					SendMessage(user.UserID, "Меню", keyboard)
				}
			}

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
