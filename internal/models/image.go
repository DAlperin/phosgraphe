package models

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	PrettyName string `gorm:"not null"`
	Namespace  string `gorm:"not null"`
	Source     string `gorm:"not null"`
	Hash       string
}
