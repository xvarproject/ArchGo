package models

import (
	"github.com/google/uuid"
	"time"
)

type BillingHistory struct {
	ID                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Operator          string    `gorm:"type:varchar(255);not null"`
	AccountEmail      string    `gorm:"not null"`
	Amount            int64     `gorm:"bigint;not null"`
	TransactionType   string    `gorm:"type:varchar(255)"`
	TransactionDetail string    `gorm:"type:varchar(255)"`
	TransactionTime   time.Time `gorm:"not null"`
}
