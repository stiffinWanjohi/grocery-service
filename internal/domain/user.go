package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	AdminRole    UserRole = "admin"
	CustomerRole UserRole = "customer"
)

type User struct {
	ID        uuid.UUID `json:"id"                gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email     string    `json:"email"             gorm:"type:varchar(255);unique;not null"`
	Password  string    `json:"-"                 gorm:"type:varchar(255)"`
	Name      string    `json:"name"              gorm:"type:varchar(255);not null"`
	Phone     string    `json:"phone"             gorm:"type:varchar(50);not null"`
	Address   string    `json:"address"           gorm:"type:text"`
	Picture   string    `json:"picture,omitempty" gorm:"type:text"`
	Role      UserRole  `json:"role"              gorm:"type:varchar(50);default:'user'"`
	CreatedAt time.Time `json:"created_at"        gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at"        gorm:"not null;default:current_timestamp"`
	Tokens    []Token   `json:"-"                 gorm:"foreignKey:UserID"`
}

type UserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
