package service

import (
	"errors"
	"github.com/ZnNr/user-reward-controller/internal/models"
	"github.com/ZnNr/user-reward-controller/internal/repository"

	"github.com/google/uuid"
	"time"

	"go.uber.org/zap"
)

type ReferralService struct {
	repo      repository.ReferralRepository
	logger    *zap.Logger
	referrals map[string]models.Referral // Изменено на структуру Referral
}

func NewReferralService(repo repository.ReferralRepository, logger *zap.Logger) *ReferralService {
	return &ReferralService{
		repo:      repo,
		logger:    logger,
		referrals: make(map[string]models.Referral), // Инициализация карты
	}
}

func (s *ReferralService) CreateReferral(userID string, code string) (*models.Referral, error) {
	// Проверка на дубликат
	for _, referral := range s.referrals {
		if referral.Code == code {
			return nil, errors.New("referral code already exists")
		}
	}

	referralID := generateReferralID()
	newReferral := models.Referral{
		ReferralID: referralID,
		UserID:     userID,
		Code:       code,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	s.referrals[referralID] = newReferral
	s.logger.Info("Created new referral", zap.String("referralID", referralID), zap.String("userID", userID), zap.String("code", code))
	return &newReferral, nil
}

// GetReferral получает реферальный код по ID
func (s *ReferralService) GetReferral(referralID string) (*models.Referral, error) {
	referral, found := s.referrals[referralID]
	if !found {
		return nil, errors.New("referral not found")
	}
	return &referral, nil
}

// GetReferralsByUserID получает все рефералы для пользователя
func (s *ReferralService) GetReferralsByUserID(userID string) ([]models.Referral, error) {
	var userReferrals []models.Referral
	for _, referral := range s.referrals {
		if referral.UserID == userID {
			userReferrals = append(userReferrals, referral)
		}
	}
	return userReferrals, nil
}

func (s *ReferralService) UpdateReferral(referralID string, code string) (*models.Referral, error) {
	referral, found := s.referrals[referralID]
	if !found {
		return nil, errors.New("referral not found")
	}
	if referral.Code == code {
		return &referral, nil // Код не изменился
	}
	referral.Code = code
	referral.UpdatedAt = time.Now()
	s.referrals[referralID] = referral
	s.logger.Info("Updated referral", zap.String("referralID", referralID), zap.String("newCode", code))
	return &referral, nil
}

// DeleteReferral удаляет реферальный код по ID
func (s *ReferralService) DeleteReferral(referralID string) error {
	_, found := s.referrals[referralID]
	if !found {
		return errors.New("referral not found")
	}
	delete(s.referrals, referralID)
	return nil
}

// Генерация уникального ID для реферального кода
func generateReferralID() string {

	return uuid.New().String()
}

// ValidateReferralCode проверяет, действителен ли реферальный код
func (s *ReferralService) ValidateReferralCode(code string) (bool, error) {
	for _, referral := range s.referrals {
		if referral.Code == code {
			return true, nil // Код действителен
		}
	}
	return false, errors.New("invalid referral code")
}
