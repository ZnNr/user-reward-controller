package database

import (
	"database/sql"
	"errors"
	"github.com/ZnNr/user-reward-controler/internal/models"
)

const (
	CreateReferralQuery       = `INSERT INTO referral (UserID, Code, CreatedAt, UpdatedAt) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING ID, UserID, Code, CreatedAt, UpdatedAt`
	GetReferralQuery          = `SELECT ID, UserID, Code, CreatedAt, UpdatedAt FROM referral WHERE ID = ?`
	GetReferralsByUserIDQuery = `SELECT ID, UserID, Code, CreatedAt, UpdatedAt FROM referral WHERE UserID = ?`
	UpdateReferralQuery       = `UPDATE referral SET Code = ?, UpdatedAt = CURRENT_TIMESTAMP WHERE ID = ? RETURNING ID, UserID, Code, CreatedAt, UpdatedAt`
	DeleteReferralQuery       = `DELETE FROM referral WHERE ID = ?`
)

type ReferralRepository struct {
	db *sql.DB
}

// NewReferralRepository создает новый экземпляр ReferralRepository
func NewReferralRepository(db *sql.DB) *ReferralRepository {
	return &ReferralRepository{db: db}
}

// CreateReferral создает новый реферальный код для пользователя
func (r *ReferralRepository) CreateReferral(userID string, code string) (*models.Referral, error) {
	referral := &models.Referral{}

	err := r.db.QueryRow(CreateReferralQuery, userID, code).Scan(&referral.UserID, &referral.UserID, &referral.Code, &referral.CreatedAt, &referral.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return referral, nil
}

// GetReferral возвращает реферал по его ID
func (r *ReferralRepository) GetReferral(referralID string) (*models.Referral, error) {
	referral := &models.Referral{}

	row := r.db.QueryRow(GetReferralQuery, referralID)

	err := row.Scan(&referral.UserID, &referral.UserID, &referral.Code, &referral.CreatedAt, &referral.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("referral not found")
		}
		return nil, err
	}

	return referral, nil
}

// GetReferralsByUserID возвращает список рефералов для указанного пользователя
func (r *ReferralRepository) GetReferralsByUserID(userID string) ([]models.Referral, error) {

	rows, err := r.db.Query(GetReferralsByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	referrals := make([]models.Referral, 0)
	for rows.Next() {
		referral := models.Referral{}
		err := rows.Scan(&referral.UserID, &referral.UserID, &referral.Code, &referral.CreatedAt, &referral.UpdatedAt)
		if err != nil {
			return nil, err
		}
		referrals = append(referrals, referral)
	}

	return referrals, nil
}

// UpdateReferral обновляет указанный реферальный код
func (r *ReferralRepository) UpdateReferral(referralID string, code string) (*models.Referral, error) {

	referral := &models.Referral{}

	err := r.db.QueryRow(UpdateReferralQuery, code, referralID).Scan(&referral.UserID, &referral.UserID, &referral.Code, &referral.CreatedAt, &referral.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("referral not found")
		}
		return nil, err
	}

	return referral, nil
}

// DeleteReferral удаляет реферальный код по его ID
func (r *ReferralRepository) DeleteReferral(referralID string) error {

	result, err := r.db.Exec(DeleteReferralQuery, referralID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("referral not found")
	}

	return nil
}
