package models

import (
	"github.com/google/uuid"
)

type RechargingCard struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id,omitempty"`
	Password       string    `gorm:"not null" json:"password,omitempty"`
	BalanceInCents int64     `gorm:"not null" json:"balance_in_cents,omitempty"`
	VendorName     string    `gorm:"type:varchar(255)" json:"vendor_name,omitempty"`
	Status         string    `gorm:"not null" json:"status,omitempty"`
}

type ChargeMyselfWithRechargingCardInput struct {
	CardID   uuid.UUID `json:"card_id" binding:"required"`
	Password string    `json:"password" binding:"required"`
}

type BulkCreateRechargingCardsInput struct {
	Amount         int64 `json:"amount" binding:"required"`
	BalanceInCents int64 `json:"balance_in_cents" binding:"required"`
}
