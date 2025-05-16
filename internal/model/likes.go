package model

import (
	"gorm.io/gorm"
)

type UserLikeShip struct {
	gorm.Model
	User      string `gorm:"user"`
	LikedUser string `gorm:"liked_user"`
}

type PostLikeShip struct {
	gorm.Model
	User      string `gorm:"user"`
	LikedPost int    `gorm:"liked_post"`
}

type CommentLikeShip struct {
	gorm.Model
	User         string `gorm:"user"`
	LikedComment int    `gorm:"liked_comment"`
}

func LikeUserShip(user string, likedUser string) error {
	err := DB.Model(&UserLikeShip{}).Create(&UserLikeShip{
		User:      user,
		LikedUser: likedUser,
	}).Error
	return err
}

func UnlikeUserShip(user string, likedUser string) error {
	err := DB.Model(&UserLikeShip{}).Where("user=? AND liked_user=?", user, likedUser).Delete(&UserLikeShip{}).Error
	return err
}

func LikePostShip(user string, likedPost int) error {
	err := DB.Model(&PostLikeShip{}).Create(&PostLikeShip{
		User:      user,
		LikedPost: likedPost,
	}).Error
	return err
}

func UnlikePostShip(user string, likedPost int) error {
	err := DB.Model(&PostLikeShip{}).Where("user=? AND liked_post=?", user, likedPost).Delete(&PostLikeShip{}).Error
	return err
}

func HasLikeUserShip(user string, likedUser string) (bool, error) {
	var count int64
	err := DB.Model(&UserLikeShip{}).Where("user=? AND liked_user=?", user, likedUser).Count(&count).Error
	return count > 0, err
}

func HasLikePostShip(user string, likedPost int) (bool, error) {
	var count int64
	err := DB.Model(&PostLikeShip{}).Where("user=? AND liked_post=?", user, likedPost).Count(&count).Error
	return count > 0, err
}

func HasLikeCommentShip(user string, likedComment int) (bool, error) {
	var count int64
	err := DB.Model(&CommentLikeShip{}).Where("user=? AND liked_comment=?", user, likedComment).Count(&count).Error
	return count > 0, err
}
