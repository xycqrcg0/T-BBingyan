package model

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	Name       string `gorm:"name;unique"` //用name区分
	Password   string `gorm:"password"`
	AddedAdmin string `gorm:"added_admin"`
}

func CreateAdmin(admin *Admin) error {
	err := DB.Model(&Admin{}).Create(admin).Error
	return err
}

func DeleteAdmin(name string) error {
	err := DB.Model(&Admin{}).Where("name=?", name).Delete(&Admin{}).Error
	return err
}

func HasAdmin(name string) (*Admin, error) {
	admin := &Admin{}
	result := DB.Model(&Admin{}).Where("name=?", name).First(admin)
	return admin, result.Error
}
