package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email     string `gorm:"email;not null;unique"`
	Name      string `gorm:"name;not null"`
	Password  string `gorm:"password;not null"`
	Signature string `gorm:"signature;default:你好，世界"`
	Likes     int    `gorm:"likes"`
	Follows   int    `gorm:"follows"`
}

func AddUser(newUser *User) error {
	err := DB.Model(&User{}).Create(newUser).Error
	return err
}

func DeleteUser(email string) error {
	err := DB.Model(&User{}).Where("email=?", email).Delete(&User{}).Error
	return err
}

func UpdateUser(user *User) error {
	err := DB.Model(&User{}).Where("email=?", user.Email).Updates(user).Error
	return err
}

func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := DB.Model(&User{}).Where("email=?", email).First(user).Error
	return user, err
}

func GetAllUsersInfo(page int, pageSize int) ([]User, error) {
	users := make([]User, 0)
	err := DB.Model(&User{}).Limit(pageSize).Offset(page * pageSize).Find(&users).Error

	return users, err
}

func GetUserLikes(user string) (int, error) {
	var likes int
	err := DB.Model(&User{}).Select("likes").Where("email=?", user).First(&likes).Error
	return likes, err
}
