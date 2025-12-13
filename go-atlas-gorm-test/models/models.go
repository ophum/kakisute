package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name  string `gorm:"size:255;not null"`
	Email string `gorm:"size:255;uniqueIndex;not null"`
	Posts []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
	gorm.Model
	Title   string `gorm:"not null"`
	Content string `gorm:"type:text"`
	UserID  uint   `gorm:"not null"`
	User    User   `gorm:"foreignKey:UserID"`
}
