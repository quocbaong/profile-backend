package models

import (
	"time"
)

// User represents the admin user for authentication
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"` // omit password in JSON responses
	Name      string    `gorm:"size:150" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project represents a portfolio project
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Category    string    `gorm:"size:100" json:"category"`
	Description string    `gorm:"type:text" json:"description"`
	Tags        string    `gorm:"size:255" json:"tags"`
	Image       string    `gorm:"size:500" json:"image"`
	LiveURL     string    `gorm:"size:500" json:"live_url"`
	GithubURL   string    `gorm:"size:500" json:"github_url"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContactMessage represents messages sent from the portfolio contact form
type ContactMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	LastName  string    `gorm:"size:100" json:"lastname"`
	Email     string    `gorm:"size:150;not null" json:"email"`
	Subject   string    `gorm:"size:200" json:"subject"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// ProfileSettings represents personal profile information and social links
type ProfileSettings struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:150;not null" json:"name"`
	Title            string    `gorm:"size:200;not null" json:"title"`
	Subtitle         string    `gorm:"size:200" json:"subtitle"`
	AboutTitle       string    `gorm:"size:250" json:"about_title"`
	AboutDescription string    `gorm:"type:text" json:"about_description"`
	Email            string    `gorm:"size:150" json:"email"`
	Phone            string    `gorm:"size:50" json:"phone"`
	Address          string    `gorm:"size:250" json:"address"`
	HeroImage        string    `gorm:"size:500" json:"hero_image"`
	AboutImage       string    `gorm:"size:500" json:"about_image"`
	ResumeURL        string    `gorm:"size:500" json:"resume_url"`
	GithubURL        string    `gorm:"size:500" json:"github_url"`
	LinkedinURL      string    `gorm:"size:500" json:"linkedin_url"`
	FacebookURL      string    `gorm:"size:500" json:"facebook_url"`
	InstagramURL     string    `gorm:"size:500" json:"instagram_url"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ServiceItem represents a service offering
type ServiceItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	IconClass   string    `gorm:"size:100" json:"icon_class"`
	Description string    `gorm:"type:text" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SkillItem represents a technology or tool skill
type SkillItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Category  string    `gorm:"size:100" json:"category"`
	Color     string    `gorm:"size:50" json:"color"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
