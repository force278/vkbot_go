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
	_ "github.com/lib/pq"
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
		fmt.Printf("Unable to connect to database: %v\n", err)
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
	// Создаем таблицы, если они не существуют
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
			last_message TIMESTAMP,
			state INT DEFAULT 0
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
		_, err := DB.Exec(context.Background(), table)
		if err != nil {
			fmt.Printf("Ошибка при создании таблицы: %v\n", err)
			return err
		}
	}
	fmt.Println("Все необходимые таблицы созданы или уже существуют")
	return nil
}

func AddStatusColumnIfNotExists() error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	// Проверяем, существует ли колонка 'status' в таблице 'bibinto'
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='bibinto' AND column_name='status'
		);
	`
	err := DB.QueryRow(context.Background(), query).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка при проверке существования колонки: %v\n", err)
		return err
	}
	// Если колонка не существует, добавляем ее
	if !exists {
		alterQuery := `ALTER TABLE bibinto ADD COLUMN status int;`
		_, err := DB.Exec(context.Background(), alterQuery)
		if err != nil {
			fmt.Printf("Ошибка при добавлении колонки: %v\n", err)
			return err
		}
		fmt.Println("Колонка 'status' успешно добавлена в таблицу 'bibinto'")
	} else {
		fmt.Println("Колонка 'status' уже существует в таблице 'bibinto'")
	}
	return nil
}

// Возвращает пользователя, флаг - есть ли юзер в бд и ошибку
func GetUser(userid uint) (utils.User, bool, error) {
	if DB == nil {
		return utils.User{}, false, fmt.Errorf("database connection is not established")
	}
	query := fmt.Sprintf("SELECT * FROM bibinto where userid = %d", userid) // Предположим, у вас есть таблица users
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Ошибка выполения запроса GetUser: %v\n", err)
	}
	defer rows.Close() // Закрываем rows после завершения работы с ними
	var user utils.User
	var Score, People, Active, Ban, Admin, Address, Sub, State sql.NullInt32
	var Name, Photo sql.NullString
	var LastMessage sql.NullTime
	// Проверяем, есть ли строки в результате
	if rows.Next() {
		// Если строки есть, сканируем их
		if err := rows.Scan(&user.ID, &user.UserID, &Name, &Photo, &Score, &People, &Active, &Ban, &Admin, &Address, &Sub, &LastMessage, &State); err != nil {
			return utils.User{}, false, fmt.Errorf("failed to scan row: %v", err)
		}
		if Name.Valid {
			user.Name = Name.String
		}
		if Photo.Valid {
			user.Photo = Photo.String
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
		return utils.User{}, false, fmt.Errorf("user with userid %d not found", userid)
	}
	// Проверка на ошибки после обхода строк
	if err := rows.Err(); err != nil {
		return utils.User{}, false, err
	}
	return user, true, nil
}

func AddUser(userid uint) (uint, error) {
	var user utils.User
	user.UserID = userid
	if DB == nil {
		return 0, fmt.Errorf("database connection is not established")
	}
	query := `
		INSERT INTO bibinto (user_id, name, photo, score, people, ban, address, admin, sub, last_message, state)
		VALUES (\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11)
		RETURNING id
	`
	// Выполнение запроса
	var id uint
	err := DB.QueryRow(context.Background(), query, user.UserID, user.Name, user.Photo, user.Score, user.People, user.Ban, user.Address, user.Admin, user.Sub, user.LastMessage, user.State).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	return id, nil
}

func UpdateUser(user utils.User) error {
	// Проверяем соединение с базой данных
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	// Формируем SQL-запрос
	query := `
		UPDATE bibinto SET
			name = COALESCE(NULLIF(\$1, ''), name),
			photo = COALESCE(NULLIF(\$2, ''), photo),
			score = COALESCE(NULLIF(\$3, -1), score),
			people = COALESCE(NULLIF(\$4, -1), people),
			ban = COALESCE(\$5, ban),
			address = COALESCE(\$6, address),
			admin = COALESCE(\$7, admin),
			sub = COALESCE(\$8, sub),
			last_message = COALESCE(\$9, last_message)
			state = COALESCE(\$10, state)
		WHERE id = \$11
	`

	// Выполнение запроса
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
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func GetRec(userid uint) (utils.User, error) {
	if DB == nil {
		fmt.Print("Database connection is not established")
	}
	query := fmt.Sprintf("SELECT * FROM bibinto WHERE UserID = (SELECT UserID FROM stack WHERE UserID NOT IN (SELECT UserID FROM history WHERE ValuerID = %d) AND UserID <> %d ORDER BY id DESC LIMIT 1)", userid, userid) // Предположим, у вас есть таблица users
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Ошибка выполения запроса GetRec: %v\n", err)
	}
	defer rows.Close() // Закрываем rows после завершения работы с ними
	var user utils.User
	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.UserID, &user.Name, &user.Photo, &user.Score, &user.People, &user.Active, &user.Ban, &user.Admin, &user.Address, &user.Sub, &user.LastMessage, &user.State); err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
		}
	}
	// Проверка на ошибки после обхода строк
	if err := rows.Err(); err != nil {
		fmt.Printf("Error occurred during rows iteration: %v\n", err)
	}
	return user, nil
}

func AddStack(userid uint) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	query := `
		INSERT INTO stack (userid)
		VALUES (\$1)
		RETURNING id
	`
	// Выполнение запроса
	var id uint
	err := DB.QueryRow(context.Background(), query, userid).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert stack: %w", err)
	}

	return nil
}

func DeleteHistory(userid uint) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `DELETE FROM bibinto WHERE id = \$1` // Параметризованный запрос для предотвращения SQL-инъекций

	// Выполняем запрос на удаление
	result, err := DB.Exec(context.Background(), query, userid)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса Delete:User  %v\n", err)
		return err
	}

	// Проверяем, было ли удалено хотя бы одно значение
	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userid)
	}
	return nil
}

func Ban(userid uint) error {
	// Проверяем соединение с базой данных
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	// Формируем SQL-запрос
	query := `
		UPDATE bibinto SET
			ban = \$1
		WHERE id = \$2
	`

	// Выполнение запроса
	_, err := DB.Exec(context.Background(), query,
		1,
		userid,
	)
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}
	return nil
}

func AddGrade(userid uint, valuerid uint, grade int, message *string) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}

	query := `INSERT INTO grades (user_id, valuerid, grade, message) VALUES (\$1, \$2, \$3, \$4)`
	// Выполняем запрос на добавление оценки
	_, err := DB.Exec(context.Background(), query, userid, valuerid, grade, message)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса AddRating: %v\n", err)
		return err
	}

	return nil
}

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
WHERE grades.userid = \$1 
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
			deleteQuery += fmt.Sprintf("$%d", i+1) // Параметры начинаются с \$1
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

func Top() ([]utils.User, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	// Выполняем запрос для получения топ 10 пользователей по score
	query := `SELECT * FROM bibinto ORDER BY People DESC LIMIT 3`
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса на получение топа пользователей: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var topUsers []utils.User
	for rows.Next() {
		var user utils.User
		if err := rows.Scan(&user.ID, &user.UserID, &user.Name, &user.Photo, &user.Score, &user.People, &user.Active, &user.Ban, &user.Admin, &user.Address, &user.Sub, &user.LastMessage, &user.State); err != nil {
			fmt.Printf("Ошибка при сканировании строки: %v\n", err)
			return nil, err
		}
		topUsers = append(topUsers, user)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Ошибка при итерации по строкам: %v\n", err)
		return nil, err
	}

	return topUsers, nil
}

func Top10() ([]string, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	query := "SELECT photo FROM bibinto WHERE id > (SELECT COUNT(id) FROM bibinto)-500 AND people >= 50 AND ban != 1 ORDER BY score/people DESC LIMIT 10;"
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса на получение топа пользователей: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var topPhotos []string
	for rows.Next() {
		var photo string
		if err := rows.Scan(&photo); err != nil {
			fmt.Printf("Ошибка при сканировании строки: %v\n", err)
			return nil, err
		}
		topPhotos = append(topPhotos, photo)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Ошибка при итерации по строкам: %v\n", err)
		return nil, err
	}

	return topPhotos, nil
}

func MyTop(userid uint) (int, error) {
	if DB == nil {
		return -1, fmt.Errorf("database connection is not established")
	}

	query := fmt.Sprintf("SELECT COUNT(u2.id) FROM users u1 LEFT JOIN users u2 ON u2.people > u1.people OR (u2.people = u1.people AND u2.id <= u1.id) WHERE u1.UserID = %d", userid)
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса на получение моего топа: %v\n", err)
		return -1, err
	}
	defer rows.Close()

	var result int
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			fmt.Printf("Ошибка при сканировании строки: %v\n", err)
			return -1, err
		}
	}

	return result, nil
}

func WasUser(userid uint) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database connection is not established")
	}

	query := fmt.Sprintf("SELECT EXISTS(SELECT * FROM users WHERE UserID = %d)", userid)
	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса WasUser: %v\n", err)
		return false, err
	}
	defer rows.Close()

	var result int
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			fmt.Printf("Ошибка при сканировании строки: %v\n", err)
			return false, err
		}
	}

	return result != 0, nil
}

func IsFull(userid uint) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database connection is not established")
	}
	var user utils.User

	// SQL-запрос для поиска пользователя по userID
	query := `SELECT id, name, photo FROM bibinto WHERE userID = \$1`
	err := DB.QueryRow(context.Background(), query, userid).Scan(&user.ID, &user.Name, &user.Photo)
	if err != nil {
		if err == sql.ErrNoRows {
			// Пользователь не найден
			return false, nil
		}
		// Ошибка при выполнении запроса
		return false, err
	}

	// Проверка наличия имени и фотографии
	if user.Name == "" || user.Photo == "" {
		return false, nil // Имя или фотография отсутствуют
	}

	return true, nil // Пользователь найден, имя и фотография присутствуют
}

func UpdateLastMessage(userid uint) error {
	if DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	// Получаем текущее время
	currentTime := time.Now()

	// SQL-запрос для обновления поля LastMessage
	query := `UPDATE bibinto SET LastMessage = \$1 WHERE userID = \$2`
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
	query := `UPDATE bibinto SET State = \$1 WHERE userID = \$2`
	_, err := DB.Exec(context.Background(), query, state, userid)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при обновлении
	}

	return nil // Успешное обновление
}
