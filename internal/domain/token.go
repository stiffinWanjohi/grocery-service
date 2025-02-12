package domain

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID         uuid.UUID  `json:"id"                   gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID     uuid.UUID  `json:"user_id"              gorm:"type:uuid;not null;index"`
	Token      string     `json:"token"                gorm:"type:text;not null;unique;index"`
	Type       TokenType  `json:"type"                 gorm:"type:varchar(20);not null"`
	ExpiresAt  time.Time  `json:"expires_at"           gorm:"not null;index"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	Provider   string     `json:"provider"             gorm:"type:varchar(50);not null;default:'google'"`
	ProviderID string     `json:"provider_id"          gorm:"type:varchar(255)"`
	CreatedAt  time.Time  `json:"created_at"           gorm:"not null;default:current_timestamp"`
	UpdatedAt  time.Time  `json:"updated_at"           gorm:"not null;default:current_timestamp"`
	User       *User      `json:"-"                    gorm:"foreignKey:UserID"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeID      TokenType = "id"
)

func (Token) TableName() string {
	return "tokens"
}
