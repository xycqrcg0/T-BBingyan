package model

import "gorm.io/gorm"

//因为节点是可以变的，这里就把它放数据库里吧

type Tag struct {
	gorm.Model
	Tag        string `gorm:"tag;unique"`
	AddedAdmin string `gorm:"added_admin"`
}

//ps:节点这种东西能删吗？删了对应的文章怎么办

func CreateTag(tag *Tag) error {
	err := DB.Model(&Tag{}).Create(tag).Error
	return err
}

func CheckTag(tag string) (bool, error) {
	var count int64
	err := DB.Model(&Tag{}).Where("tag=?", tag).Count(&count).Error
	return count > 0, err
}
