package models

import "time"

// Referral представляет модель данных реферального кода
type Referral struct {
	ReferralID string    `json:"referralId"` // Уникальный идентификатор реферального кода
	UserID     string    `json:"userId"`     // Идентификатор пользователя, которому принадлежит код
	Code       string    `json:"code"`       // Сам реферальный код
	CreatedAt  time.Time `json:"createdAt"`  // Дата создания реферального кода
	UpdatedAt  time.Time `json:"updatedAt"`  // Дата последнего обновления информации о реферальном коде
}

// CreateReferralRequest представляет запрос для создания нового реферального кода
type CreateReferralRequest struct {
	UserID string `json:"userId"` // Идентификатор пользователя, которому будет принадлежать код
	Code   string `json:"code"`   // Сам реферальный код
}

// UpdateReferralRequest представляет запрос для обновления существующего реферального кода
type UpdateReferralRequest struct {
	ReferralID string `json:"referralId"` // Уникальный идентификатор реферального кода
	Code       string `json:"code"`       // Обновленный реферальный код
}
