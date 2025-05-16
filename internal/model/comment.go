package model

import (
	"gorm.io/gorm"
)

// comment 只留一个主评论，其他的回复都是子评论
var (
	TIMEDESC  = "1"
	REPLYDESC = "3"
)

type Comment struct {
	gorm.Model
	Uid     string `gorm:"uid"`
	Pid     int    `gorm:"pid"`
	Root    int    `gorm:"root"`
	Parent  int    `gorm:"parent"`
	Content string `gorm:"content"`
	Replies int    `gorm:"replies"`
	Likes   int    `gorm:"likes"`
	User    User   `gorm:"foreignKey:Uid;references:Email"`
}

func CreateComment(comment *Comment) error {
	result := DB.Model(&Comment{}).Create(comment)
	return result.Error
}

func DeleteComment(id int) error {
	result := DB.Model(&Comment{}).Where("id=?", id).Delete(&Comment{})
	return result.Error
}

func DeleteRepliesByRoot(root int) (int64, error) {
	result := DB.Model(&Comment{}).Where("root=?", root).Delete(&Comment{})
	return result.RowsAffected, result.Error
}

func DeleteCommentsByPost(pid int) (int64, error) {
	result := DB.Model(&Comment{}).Where("pid=?", pid).Delete(&Comment{})
	return result.RowsAffected, result.Error
}

func ChangeCommentReplies(id int, add bool) error {
	var err error
	if add {
		err = DB.Model(&Comment{}).Where("id=?", id).Update("replies", gorm.Expr("replies+1")).Error
	} else {
		err = DB.Model(&Comment{}).Where("id=?", id).Update("replies", gorm.Expr("replies-1")).Error
	}
	return err
}

func GetCommentByUid(id int, uid string) (*Comment, error) {
	var comment Comment
	result := DB.Model(&Comment{}).Where("id=? AND uid=?", id, uid).First(&comment)
	return &comment, result.Error
}

func GetCommentByPid(id int, pid int) (*Comment, error) {
	var comment Comment
	result := DB.Model(&Comment{}).Where("id=? AND pid=?", id, pid).First(&comment)
	return &comment, result.Error
}

func GetCommentById(id int) (*Comment, error) {
	var comment Comment
	result := DB.Model(&Comment{}).Where("id=?", id).First(&comment)
	return &comment, result.Error
}

func GetReplyById(id int, root int) (*Comment, error) {
	var reply Comment
	result := DB.Model(&Comment{}).Where("id=? AND root=?", id, root).First(&reply)
	return &reply, result.Error
}

func GetCommentLikes(id int) (int, error) {
	var likes int
	err := DB.Model(&Comment{}).Select("likes").Where("id=?", id).First(&likes).Error
	return likes, err
}

func GetCommentsByPost(pid int, page int, pageSize int, time bool) ([]Comment, error) {
	comments := make([]Comment, 0)
	var err error
	if time {
		err = DB.Model(&Comment{}).Preload("User").Where("pid=? AND root=?", pid, 0).
			Order("created_at DESC").Limit(pageSize).Offset(page * pageSize).Find(&comments).Error
	} else {
		err = DB.Model(&Comment{}).Preload("User").Where("pid=? AND root=?", pid, 0).
			Order("replies DESC").Limit(pageSize).Offset(page * pageSize).Find(comments).Error
	}
	return comments, err
}

func GetRepliesByRoot(root int, page int, pageSize int, ty string) ([]Comment, error) {
	comments := make([]Comment, 0)
	var err error
	switch ty {
	case TIMEDESC:
		err = DB.Model(&Comment{}).Preload("User").Where("root=?", root).
			Order("created_at DESC").Limit(pageSize).Offset(page * pageSize).Find(&comments).Error
	case REPLYDESC:
		err = DB.Model(&Comment{}).Preload("User").Where("root=?", root).
			Order("replies DESC").Limit(pageSize).Offset(page * pageSize).Find(&comments).Error
	}

	return comments, err
}
