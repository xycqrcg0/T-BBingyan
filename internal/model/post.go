package model

import (
	"BBingyan/internal/global"
	"gorm.io/gorm"
)

//id是能暴露的吗

type Post struct {
	gorm.Model
	Author  string `gorm:"author;not null"`
	Title   string `gorm:"title;not null"`
	Tag     string `gorm:"tag"`
	Content string `gorm:"content;not null"`
	Likes   int    `gorm:"likes"`
	Replies int    `gorm:"replies"`
	User    User   `gorm:"user;foreignKey:Author;references:Email"`
}

func AddPost(newPost *Post) error {
	err := DB.Model(&Post{}).Create(newPost).Error
	return err
}

func DeletePostOnlyById(id int) error {
	result := DB.Model(&Post{}).Where("id=?", id).Delete(&Post{})
	if result.RowsAffected == 0 {
		return global.ErrPostNone
	}
	return result.Error
}

func DeletePostById(user string, id int) error {
	result := DB.Model(&Post{}).Where("id=? AND author=?", id, user).Delete(&Post{})
	if result.RowsAffected == 0 {
		return global.ErrPostNone
	}
	return result.Error
}

func ChangePostReplies(id int, add bool, num int) error {
	var err error
	if add {
		err = DB.Model(&Post{}).Where("id=?", id).Update("replies", gorm.Expr("replies+?", num)).Error
	} else {
		err = DB.Model(&Post{}).Where("id=?", id).Update("replies", gorm.Expr("replies-?", num)).Error
	}
	return err
}

func GetPostById(id int) (*Post, error) {
	post := &Post{}
	err := DB.Model(&Post{}).Preload("User").Where("id=?", id).First(post).Error
	return post, err
}

func GetPostLikes(id int) (int, error) {
	var likes int
	err := DB.Model(&Post{}).Select("likes").Where("id=?", id).First(&likes).Error
	return likes, err
}

func GetPostsByEmail(email string, page int, pageSize int) ([]Post, error) {
	posts := make([]Post, 0)
	err := DB.Model(&Post{}).Preload("User").Where("author=?", email).
		Limit(pageSize).Offset(pageSize * page).Find(&posts).Error
	return posts, err
}

func GetPostsByTagTime(tag string, page int, pageSize int, desc bool) ([]Post, error) {
	posts := make([]Post, 0)
	var err error
	if desc {
		err = DB.Model(&Post{}).Preload("User").Where("tag=?", tag).
			Order("created_at DESC").Limit(pageSize).Offset(page * pageSize).Find(&posts).Error
	} else {
		err = DB.Debug().Model(&Post{}).Preload("User").Where("tag=?", tag).
			Order("created_at").Limit(pageSize).Offset(pageSize * page).Find(&posts).Error
	}
	return posts, err
}

func GetPostsByTagReplies(tag string, page int, pageSize int, desc bool) ([]Post, error) {
	posts := make([]Post, 0)
	var err error

	if desc {
		err = DB.Model(&Post{}).Preload("User").Where("tag=?", tag).
			Order("replies DESC").Limit(pageSize).Offset(pageSize * page).Find(&posts).Error
	} else {
		err = DB.Model(&Post{}).Preload("User").Where("tag=?", tag).
			Order("replies").Limit(pageSize).Offset(pageSize * page).Find(&posts).Error
	}

	return posts, err
}
