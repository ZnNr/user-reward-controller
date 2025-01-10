package repository

import "github.com/ZnNr/user-reward-controler/internal/models"

type ReferralRepository interface {
	CreateReferral(userID string, code string) (*models.Referral, error)
	GetReferral(referralID string) (*models.Referral, error)
	GetReferralsByUserID(userID string) ([]models.Referral, error)
	UpdateReferral(referralID string, code string) (*models.Referral, error)
	DeleteReferral(referralID string) error
}
