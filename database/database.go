package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
	"vkbot/config"
	"vkbot/utils"

	"github.com/jackc/pgx/v4/pgxpool"
)

// DB представляет собой пул соединений с базой данных
var DB *pgxpool.Pool

// Connect создает соединение с базой данных PostgreSQL
func Connect() {
	var err error
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.AppConfig.DbUser,
		config.AppConfig.DbPassword,
		config.AppConfig.DbHost,
		config.AppConfig.DbPort,
		config.AppConfig.DbName,
	)

	DB, err = pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	fmt.Println("Successfully connected to the database")
}

// Disconnect закрывает соединение с базой данных
func Disconnect() {
	if DB != nil {
		DB.Close()
		fmt.Println("Database connection closed")
	}
}

// CreateDatabaseAndTables создает базу данных и необходимые таблицы, если они не существуют
func CreateDatabaseAndTables() error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS bibinto (
			id SERIAL PRIMARY KEY,
			userid BIGINT NOT NULL,
			name TEXT,
			photo TEXT,
			score INT DEFAULT 0,
			people INT DEFAULT 0,
			active BOOLEAN DEFAULT TRUE,
			ban BOOLEAN DEFAULT FALSE,
			admin BOOLEAN DEFAULT FALSE,
			address TEXT,
			sub TEXT,
			lastmessage TIMESTAMP,
			state INT DEFAULT 0,
			recuser BIGINT DEFAULT 0,
			recmess TEXT,
			about TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS stack (
			id SERIAL PRIMARY KEY,
			userid BIGINT NOT NULL,
			FOREIGN KEY (userid) REFERENCES bibinto(userid) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS grades (
			id SERIAL PRIMARY KEY,
			userid BIGINT NOT NULL,
			valuerid BIGINT NOT NULL,
			grade INT NOT NULL,
			message TEXT,
			FOREIGN KEY (userid) REFERENCES bibinto(userid) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS history (
			id SERIAL PRIMARY KEY,
			userid BIGINT NOT NULL,
			valuerid BIGINT NOT NULL,
			FOREIGN KEY (userid) REFERENCES bibinto(userid) ON DELETE CASCADE
		);`,
	}

	for _, table := range tables {
		if _, err := DB.Exec(context.Background(), table); err != nil {
			log.Printf("Ошибка при создании таблицы: %v\n", err)
			return err
		}
	}
	fmt.Println("Все необходимые таблицы созданы или уже существуют")
	return nil
}

// GetUser  возвращает пользователя, флаг - есть ли юзер в бд и ошибку
func GetUser(userid uint) (utils.User, bool, error) {
	if DB == nil {
		return utils.User{}, false, fmt.Errorf("database connection is not established")
	}

	query := `SELECT * FROM bibinto WHERE userid = $1`
	rows, _ := DB.Query(context.Background(), query, userid)

	var user utils.User
	var Score, People, Active, Ban, Admin, Address, Sub, State sql.NullInt32
	var Name, Photo, RecMess, About sql.NullString
	var LastMessage sql.NullTime
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&user.ID, &user.UserID, &Name, &Photo, &Score, &People, &Active, &Ban, &Admin, &Address, &Sub, &LastMessage, &State, &user.RecUser, &RecMess, &About); err != nil {
			return utils.User{}, false, fmt.Errorf("failed to scan row: %v", err)
		}

		// Присваиваем значения полям структуры user
		if Name.Valid {
			user.Name = Name.String
		}
		if Photo.Valid {
			user.Photo = Photo.String
		}
		if RecMess.Valid {
			user.RecMess = RecMess.String
		}
		if RecMess.Valid {
			user.About = About.String
		}
		if Score.Valid {
			user.Score = int(Score.Int32)
		}
		if People.Valid {
			user.People = int(People.Int32)
		}
		if Active.Valid {
			user.Active = int(Active.Int32)
		}
		if Ban.Valid {
			user.Ban = int(Ban.Int32)
		}
		if Admin.Valid {
			user.Admin = int(Admin.Int32)
		}
		if Address.Valid {
			user.Address = int(Address.Int32)
		}
		if LastMessage.Valid {
			user.LastMessage = LastMessage.Time
		}
		if State.Valid {
			user.State = int(State.Int32)
		}
	} else {
		// Если строки нет, возвращаем ошибку, что пользователь не найден
		return utils.User{}, false, nil
	}
	// Проверка на ошибки после обхода строк
	if err := rows.Err(); err != nil {
		return utils.User{}, false, err
	}

	return user, true, nil
}

// AddUser  добавляет нового пользователя в базу данных
func AddUser(userid uint) (uint, error) {
	if DB == nil {
		return 0, fmt.Errorf("database connection is not established")
	}

	query := `
	INSERT INTO bibinto (userid, name, photo, score, people, ban, address, admin, sub, lastmessage, state, recuser, recmess, about)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	RETURNING id
`
	var id uint
	err := DB.QueryRow(context.Background(), query, userid, "", "", 0, 0, 0, 0, 0, 0, time.Now(), 0, 0, "", "").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	return id, nil
}

// UpdateUser  обновляет информацию о пользователе в базе данных
func UpdateUser(user utils.User) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `
		UPDATE bibinto SET
			name = COALESCE(NULLIF($1, ''), name),
			photo = COALESCE(NULLIF($2, ''), photo),
			score = COALESCE(NULLIF($3, -1), score),
			people = COALESCE(NULLIF($4, -1), people),
			ban = COALESCE($5, ban),
			address = COALESCE($6, address),
			admin = COALESCE($7, admin),
			sub = COALESCE($8, sub),
			lastmessage = COALESCE($9, lastmessage),
			state = COALESCE($10, state),
			recuser = COALESCE($11, recuser),
			recmess = COALESCE($12, recmess),
			about = COALESCE($13, about)
		WHERE id = $14
	`

	_, err := DB.Exec(context.Background(), query,
		user.Name,
		user.Photo,
		user.Score,
		user.People,
		user.Ban,
		user.Address,
		user.Admin,
		user.Sub,
		user.LastMessage,
		user.State,
		user.RecUser,
		user.RecMess,
		user.About,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// GetRec получает следующего пользователя для оценки
func GetRec(userid uint) (utils.User, bool, error) {
	if DB == nil {
		return utils.User{}, false, fmt.Errorf("database connection is not established")
	}

	query := `
	SELECT * FROM bibinto 
	WHERE userid = (
		SELECT userid FROM stack 
		WHERE userid NOT IN (SELECT userid FROM history WHERE valuerid = $1) 
		AND userid <> $1 
		ORDER BY id DESC LIMIT 1
	)`

	row := DB.QueryRow(context.Background(), query, userid)

	var user utils.User
	var Score, People, Active, Ban, Admin, Address, Sub, State sql.NullInt32
	var Name, Photo, RecMess, About sql.NullString
	var LastMessage sql.NullTime

	if err := row.Scan(&user.ID, &user.UserID, &Name, &Photo, &Score, &People, &Active, &Ban, &Admin, &Address, &Sub, &LastMessage, &State, &user.RecUser, &RecMess, &About); err != nil {
		if err == sql.ErrNoRows {
			return utils.User{}, false, nil // Пользователь не найден
		}
		return utils.User{}, false, nil
	}

	// Присваиваем значения полям структуры user
	if Name.Valid {
		user.Name = Name.String
	}
	if Photo.Valid {
		user.Photo = Photo.String
	}
	if RecMess.Valid {
		user.RecMess = RecMess.String
	}
	if About.Valid {
		user.About = About.String
	}
	if Score.Valid {
		user.Score = int(Score.Int32)
	}
	if People.Valid {
		user.People = int(People.Int32)
	}
	if Active.Valid {
		user.Active = int(Active.Int32)
	}
	if Ban.Valid {
		user.Ban = int(Ban.Int32)
	}
	if Admin.Valid {
		user.Admin = int(Admin.Int32)
	}
	if Address.Valid {
		user.Address = int(Address.Int32)
	}
	if LastMessage.Valid {
		user.LastMessage = LastMessage.Time
	}
	if State.Valid {
		user.State = int(State.Int32)
	}

	return user, true, nil
}

// AddStack добавляет пользователя в стек
func AddStack(userid uint) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `INSERT INTO stack (userid) VALUES ($1) RETURNING id`
	_, err := DB.Exec(context.Background(), query, userid)
	if err != nil {
		return fmt.Errorf("failed to insert stack: %w", err)
	}

	return nil
}

// DeleteHistory удаляет историю пользователя
func DeleteHistory(userid uint) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `DELETE FROM history WHERE userid = $1`

	result, err := DB.Exec(context.Background(), query, userid)
	if err != nil {
		log.Printf("Ошибка выполнения запроса DeleteHistory: %v\n", err)
		return err
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("history for user with id %d not found", userid)
	}
	return nil
}

// Ban запрещает пользователя
func Ban(userid uint64) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `UPDATE bibinto SET ban = $1 WHERE userid = $2`
	_, err := DB.Exec(context.Background(), query, 1, userid)
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}
	return nil
}

// Ban запрещает пользователя
func Unban(userid uint64) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `UPDATE bibinto SET ban = $1 WHERE userid = $2`
	_, err := DB.Exec(context.Background(), query, 0, userid)
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}
	return nil
}

func AddSub(userid uint64) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `UPDATE bibinto SET Sub = $1 WHERE userid = $2`
	_, err := DB.Exec(context.Background(), query, 1, userid)
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}
	return nil
}

func PopSub(userid uint64) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `UPDATE bibinto SET Sub = $1 WHERE userid = $2`
	_, err := DB.Exec(context.Background(), query, 0, userid)
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}
	return nil
}

// AddGrade добавляет оценку пользователя
func AddGrade(userid uint, valuerid uint, grade int, message *string) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `INSERT INTO grades (userid, valuerid, grade, message) VALUES ($1, $2, $3, $4)`
	_, err := DB.Exec(context.Background(), query, userid, valuerid, grade, message)
	if err != nil {
		log.Printf("Ошибка выполнения запроса AddGrade: %vn", err)
		return err
	}

	return nil
}

// GetGrades получает оценки пользователя
func GetGrades(userid uint) ([]utils.Grade, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}
	// Начинаем транзакцию
	tx, err := DB.Begin(context.Background())
	if err != nil {
		fmt.Printf("Ошибка при начале транзакции: %v\n", err)
		return nil, err
	}
	defer tx.Rollback(context.Background()) // Откат транзакции в случае ошибки
	// Получаем пять последних оценок
	query := `SELECT grades.id, grades.userid, grades.valuerid, grades.grade, grades.message 
FROM grades 
JOIN bibinto ON grades.valuerid = bibinto.userid 
WHERE grades.userid = $1 
ORDER BY grades.message LIMIT 5`
	rows, err := tx.Query(context.Background(), query, userid)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса на получение оценок: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var grades []utils.Grade
	var ids []uint // Для хранения ID оценок, которые нужно удалить
	for rows.Next() {
		var g utils.Grade
		if err := rows.Scan(&g.ID, &g.UserID, &g.ValuerID, &g.Grade, &g.Message); err != nil {
			fmt.Printf("Ошибка при сканировании строки: %v\n", err)
			return nil, err
		}
		grades = append(grades, g)
		ids = append(ids, g.ID) // Сохраняем ID для удаления
	}

	// Удаляем полученные оценки, если они есть
	if len(ids) > 0 {
		// Формируем строку запроса на удаление с правильным количеством параметров
		deleteQuery := "DELETE FROM grades WHERE id IN ("
		for i := range ids {
			if i > 0 {
				deleteQuery += ", "
			}
			deleteQuery += fmt.Sprintf("$%d", i+1) // Параметры начинаются с $1
		}
		deleteQuery += ")"

		// Выполняем удаление оценок
		if _, err := tx.Exec(context.Background(), deleteQuery, ids); err != nil {
			fmt.Printf("Ошибка при удалении оценок: %v\n", err)
			return nil, err
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(context.Background()); err != nil {
		fmt.Printf("Ошибка при подтверждении транзакции: %v\n", err)
		return nil, err
	}

	return grades, nil
}

// Top получает топ пользователей
func Top() ([]utils.User, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	query := `SELECT * FROM bibinto ORDER BY People DESC LIMIT 3`
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		log.Printf("Ошибка выполнения запроса на получение топа пользователей: %v\n", err)
		return []utils.User{}, err
	}
	defer rows.Close()

	var Score, People, Active, Ban, Admin, Address, Sub, State sql.NullInt32
	var Name, Photo, RecMess, About sql.NullString
	var LastMessage sql.NullTime

	var topUsers []utils.User
	for rows.Next() {
		var user utils.User
		if err := rows.Scan(&user.ID, &user.UserID, &Name, &Photo, &Score, &People, &Active, &Ban, &Admin, &Address, &Sub, &LastMessage, &State, &user.RecUser, &RecMess, &About); err != nil {
			log.Printf("Ошибка при сканировании строки: %v\n", err)
			return nil, err
		}
		// Присваиваем значения полям структуры user
		if Name.Valid {
			user.Name = Name.String
		}
		if Photo.Valid {
			user.Photo = Photo.String
		}
		if RecMess.Valid {
			user.RecMess = RecMess.String
		}
		if About.Valid {
			user.About = About.String
		}
		if Score.Valid {
			user.Score = int(Score.Int32)
		}
		if People.Valid {
			user.People = int(People.Int32)
		}
		if Active.Valid {
			user.Active = int(Active.Int32)
		}
		if Ban.Valid {
			user.Ban = int(Ban.Int32)
		}
		if Admin.Valid {
			user.Admin = int(Admin.Int32)
		}
		if Address.Valid {
			user.Address = int(Address.Int32)
		}
		if LastMessage.Valid {
			user.LastMessage = LastMessage.Time
		}
		if State.Valid {
			user.State = int(State.Int32)
		}

		topUsers = append(topUsers, user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по строкам: %v\n", err)
		return nil, err
	}

	return topUsers, nil
}

// Top10 получает топ 10 пользователей по фотографиям
func Top10() ([]string, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	query := `SELECT photo FROM bibinto WHERE id > (SELECT COUNT(id) FROM bibinto)-500 AND people >= 50 AND ban != 1 ORDER BY score/people DESC LIMIT 10`
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		log.Printf("Ошибка выполнения запроса на получение топа пользователей: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var topPhotos []string
	for rows.Next() {
		var photo string
		if err := rows.Scan(&photo); err != nil {
			log.Printf("Ошибка при сканировании строки: %v\n", err)
			return nil, err
		}
		topPhotos = append(topPhotos, photo)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по строкам: %v\n", err)
		return nil, err
	}

	return topPhotos, nil
}

// MyTop возвращает позицию пользователя в топе
func MyTop(userid uint) (int, error) {
	if DB == nil {
		return -1, fmt.Errorf("database connection is not established")
	}

	query := `SELECT COUNT(u2.id) FROM bibinto u1 LEFT JOIN bibinto u2 ON u2.people > u1.people OR (u2.people = u1.people AND u2.id <= u1.id) WHERE u1.userid = $1`
	row := DB.QueryRow(context.Background(), query, userid)

	var result int
	if err := row.Scan(&result); err != nil {
		log.Printf("Ошибка выполнения запроса на получение моего топа: %v\n", err)
		return -1, err
	}

	return result, nil
}

// WasUser  проверяет, существует ли пользователь в базе данных
func WasUser(userid uint) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database connection is not established")
	}

	query := `SELECT EXISTS(SELECT 1 FROM bibinto WHERE userid = $1)`
	row := DB.QueryRow(context.Background(), query, userid)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		log.Printf("Ошибка выполнения запроса Was:User  %v\n", err)
		return false, err
	}

	return exists, nil
}

// IsFull проверяет, заполнена ли анкета пользователя
func IsFull(userid uint) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database connection is not established")
	}

	query := `SELECT name, photo FROM bibinto WHERE userid = $1`
	row := DB.QueryRow(context.Background(), query, userid)

	var name, photo string
	if err := row.Scan(&name, &photo); err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Пользователь не найден
		}
		return false, err
	}

	return name != "" && photo != "", nil // Проверяем наличие имени и фотографии
}

// UpdateLastMessage обновляет время последнего сообщения пользователя
func UpdateLastMessage(userid uint) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	currentTime := time.Now()
	query := `UPDATE bibinto SET lastmessage = $1 WHERE userid = $2`
	_, err := DB.Exec(context.Background(), query, currentTime, userid)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при обновлении
	}

	return nil // Успешное обновление
}

func UpdateState(userid uint, state int) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	// SQL-запрос для обновления поля LastMessage
	query := `UPDATE bibinto SET state = $1 WHERE userID = $2`
	_, err := DB.Exec(context.Background(), query, state, userid)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при обновлении
	}

	return nil // Успешное обновление
}

func DeleteAbout(userid uint) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	// SQL-запрос для обновления поля LastMessage
	query := `UPDATE bibinto SET about = $1 WHERE userID = $2`
	_, err := DB.Exec(context.Background(), query, "", userid)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при обновлении
	}

	return nil // Успешное обновление
}

func AddStateColumnIfNotExists() error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	// Проверяем, существует ли колонка 'state' в таблице 'bibinto'
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='bibinto' AND column_name='state'
		);
	`
	err := DB.QueryRow(context.Background(), query).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка при проверке существования колонки: %v\n", err)
		return err
	}
	// Если колонка не существует, добавляем ее
	if !exists {
		alterQuery := `ALTER TABLE bibinto ADD COLUMN state int;`
		_, err := DB.Exec(context.Background(), alterQuery)
		if err != nil {
			fmt.Printf("Ошибка при добавлении колонки: %v\n", err)
			return err
		}
		fmt.Println("Колонка 'state' успешно добавлена в таблицу 'bibinto'")
	} else {
		fmt.Println("Колонка 'state' уже существует в таблице 'bibinto'")
	}
	return nil
}

func AddRecUserColumnIfNotExists() error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	// Проверяем, существует ли колонка 'state' в таблице 'bibinto'
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='bibinto' AND column_name='recuser'
		);
	`
	err := DB.QueryRow(context.Background(), query).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка при проверке существования колонки: %v\n", err)
		return err
	}
	// Если колонка не существует, добавляем ее
	if !exists {
		alterQuery := `ALTER TABLE bibinto ADD COLUMN recuser bigint default 0;`
		_, err := DB.Exec(context.Background(), alterQuery)
		if err != nil {
			fmt.Printf("Ошибка при добавлении колонки: %v\n", err)
			return err
		}
		fmt.Println("Колонка 'recuser' успешно добавлена в таблицу 'bibinto'")
	} else {
		fmt.Println("Колонка 'recuser' уже существует в таблице 'bibinto'")
	}
	return nil
}

func AddRecMessColumnIfNotExists() error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	// Проверяем, существует ли колонка 'state' в таблице 'bibinto'
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='bibinto' AND column_name='recmess'
		);
	`
	err := DB.QueryRow(context.Background(), query).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка при проверке существования колонки: %v\n", err)
		return err
	}
	// Если колонка не существует, добавляем ее
	if !exists {
		alterQuery := `ALTER TABLE bibinto ADD COLUMN recmess TEXT;`
		_, err := DB.Exec(context.Background(), alterQuery)
		if err != nil {
			fmt.Printf("Ошибка при добавлении колонки: %v\n", err)
			return err
		}
		fmt.Println("Колонка 'recmess' успешно добавлена в таблицу 'bibinto'")
	} else {
		fmt.Println("Колонка 'recmess' уже существует в таблице 'bibinto'")
	}
	return nil
}

func AddAboutColumnIfNotExists() error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	// Проверяем, существует ли колонка 'state' в таблице 'bibinto'
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='bibinto' AND column_name='about'
		);
	`
	err := DB.QueryRow(context.Background(), query).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка при проверке существования колонки: %v\n", err)
		return err
	}
	// Если колонка не существует, добавляем ее
	if !exists {
		alterQuery := `ALTER TABLE bibinto ADD COLUMN about TEXT;`
		_, err := DB.Exec(context.Background(), alterQuery)
		if err != nil {
			fmt.Printf("Ошибка при добавлении колонки: %v\n", err)
			return err
		}
		fmt.Println("Колонка 'about' успешно добавлена в таблицу 'bibinto'")
	} else {
		fmt.Println("Колонка 'about' уже существует в таблице 'bibinto'")
	}
	return nil
}
