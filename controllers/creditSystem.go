package controllers

import (
	"ArchGo/constants"
	"ArchGo/logger"
	"ArchGo/models"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreditSystem struct {
	DB *gorm.DB
}

func NewCreditSystem(DB *gorm.DB) *CreditSystem {
	return &CreditSystem{DB}
}

func (cs *CreditSystem) GetBalanceByUserEmail(email string) (int64, error) {
	var user models.User
	err := cs.DB.First(&user, "email = ?", email).Error
	if err != nil {
		return 0, err
	}
	logger.Info("Got balance of user", email, "as", user.Balance)
	return user.Balance, nil
}

func (cs *CreditSystem) UpdateBalanceByUserEmail(email string, amount int64) (int64, error) {
	var user models.User
	err := cs.DB.First(&user, "email = ?", email).Error
	if err != nil {
		return 0, err
	}
	newBalance := user.Balance + amount
	err = cs.DB.Model(&user).Update("balance", newBalance).Error
	if err != nil {
		return 0, err
	}
	return newBalance, nil
}

func (cs *CreditSystem) rechargeUserEmailWithChargingCardId(email string, rechargingCardId uuid.UUID, rechargingCardPassword string) (int64, error) {
	var rechargingCard models.RechargingCard
	tx := cs.DB.Begin()
	err := tx.First(&rechargingCard, rechargingCardId).Error
	if err != nil {
		logger.Info("Charging card ", rechargingCardId, "does not exist")
		tx.Rollback()
		return 0, err
	}
	if rechargingCard.Status != constants.RechargingCardActive || rechargingCard.Password != rechargingCardPassword {
		logger.Warning("Charging card ", rechargingCardId, " is not active or password is incorrect")
		tx.Rollback()
		return 0, errors.New("charging card is not active or password is incorrect")
	}

	var user models.User
	err = tx.First(&user, "email = ?", email).Error
	if err != nil {
		return 0, err
	}
	newBalance := user.Balance + rechargingCard.BalanceInCents
	err = tx.Model(&user).Update("balance", newBalance).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	err = tx.Model(&rechargingCard).Update("status", constants.RechargingCardUsed).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	return newBalance, nil
}
