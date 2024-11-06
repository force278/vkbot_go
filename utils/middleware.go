package utils

import (
	"time"
)

type ReqCounter struct {
	requests  int
	newUsers  int
	grades    int
	lastReset time.Time
}

// Конструктор для ReqCounter
func NewReqCounter() *ReqCounter {
	return &ReqCounter{
		lastReset: time.Now(),
	}
}

// Метод для увеличения общего числа запросов
func (rc *ReqCounter) Augment() {
	rc.resetIfNewDay()
	rc.requests++
}

// Метод для увеличения числа новых пользователей
func (rc *ReqCounter) AugmentNewUser() {
	rc.resetIfNewDay()
	rc.newUsers++
}

// Метод для увеличения числа оценок
func (rc *ReqCounter) AugmentGrade() {
	rc.resetIfNewDay()
	rc.grades++
}

// Метод для сброса счетчиков, если наступил новый день
func (rc *ReqCounter) resetIfNewDay() {
	currentTime := time.Now()
	if currentTime.Year() != rc.lastReset.Year() ||
		currentTime.YearDay() != rc.lastReset.YearDay() {
		rc.requests = 0
		rc.newUsers = 0
		rc.grades = 0
		rc.lastReset = currentTime
	}
}

// Метод для получения текущих значений счетчиков
func (rc *ReqCounter) GetCounts() (int, int, int) {
	rc.resetIfNewDay()
	return rc.requests, rc.newUsers, rc.grades
}
