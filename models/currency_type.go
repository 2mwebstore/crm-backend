package models

import "time"

// CurrencyType is the master list of supported currencies.
// Currently the system focuses on KHR and USD but is extensible.
type CurrencyType struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"type:varchar(10);uniqueIndex;not null" json:"code"` // "USD", "KHR"
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`            // "US Dollar", "Cambodian Riel"
	Symbol    string    `gorm:"type:varchar(10)" json:"symbol"`                    // "$", "៛"
	IsBase    bool      `gorm:"default:false" json:"is_base"`                      // true = USD (base currency)
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedByID uint      `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
