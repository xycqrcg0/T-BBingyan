package util

import (
	"BBingyan/internal/log"
	"BBingyan/internal/model"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

var (
	LIKE       = "1"
	NEEDUNLIKE = "-1"
)

//复制粘贴诶嘿

func Archive() error {
	//把redis里的点赞信息往postgres里写

	//user
	userLikeKeys, err := model.RedisDB.Keys("userlike:*").Result() //会阻塞？
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	userLikeValues, err := model.RedisDB.MGet(userLikeKeys...).Result()
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	l1 := len(userLikeKeys)

	userLikesKeys, err := model.RedisDB.Keys("userlikes:*").Result() //会阻塞？
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	userLikesValues, err := model.RedisDB.MGet(userLikeKeys...).Result()
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	l2 := len(userLikeKeys)

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < l1; i++ {
			if userLikeValues[i] == LIKE {
				params := strings.Split(userLikeKeys[i], ":")
				tx.Model(&model.UserLikeShip{}).Create(&model.UserLikeShip{
					User:      params[1],
					LikedUser: params[2],
				})
			} else if userLikeValues[i] == NEEDUNLIKE {
				params := strings.Split(userLikeKeys[i], ":")
				tx.Model(&model.UserLikeShip{}).Where("user=? AND liked_user=?", params[1], params[2]).Delete(&model.UserLikeShip{})
			}
		}

		for i := 0; i < l2; i++ {
			params := strings.Split(userLikesKeys[i], ":")
			tx.Model(&model.User{}).Where("email=?", params[1]).Update("likes", userLikesValues[i])
		}

		return nil
	})
	if err != nil {
		log.Errorf("Fail to finish postgres transaction,error:%v", err)
		return err
	}
	//redis里记录删了
	_, e := model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {
		for _, key := range userLikeKeys {
			pipe.Del(key)
		}
		for _, key := range userLikesKeys {
			pipe.Del(key)
		}
		return nil
	})
	if e != nil {
		log.Errorf("Fail to write redis,error:%v", err)
		return err
	}

	//post
	postLikeKeys, err := model.RedisDB.Keys("postlike:*").Result() //会阻塞？
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	postLikeValues, err := model.RedisDB.MGet(postLikeKeys...).Result()
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	l1 = len(postLikeKeys)

	postLikesKeys, err := model.RedisDB.Keys("postlikes:*").Result() //会阻塞？
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	postLikesValues, err := model.RedisDB.MGet(postLikeKeys...).Result()
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	l2 = len(postLikeKeys)

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < l1; i++ {
			if postLikeValues[i] == LIKE {
				params := strings.Split(postLikeKeys[i], ":")
				post, _ := strconv.Atoi(params[2])
				tx.Model(&model.PostLikeShip{}).Create(&model.PostLikeShip{
					User:      params[1],
					LikedPost: post,
				})
			} else if postLikeValues[i] == NEEDUNLIKE {
				params := strings.Split(postLikeKeys[i], ":")
				tx.Model(&model.PostLikeShip{}).Where("user=? AND liked_post=?", params[1], params[2]).Delete(&model.PostLikeShip{})
			}
		}

		for i := 0; i < l2; i++ {
			params := strings.Split(postLikesKeys[i], ":")
			tx.Model(&model.Post{}).Where("id=?", params[1]).Update("likes", postLikesValues[i])
		}

		return nil
	})
	if err != nil {
		log.Errorf("Fail to finish postgres transaction,error:%v", err)
		return err
	}
	//redis里记录删了
	_, e = model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {

		for _, key := range postLikeKeys {
			model.RedisDB.Del(key)
		}
		for _, key := range postLikesKeys {
			model.RedisDB.Del(key)
		}
		return nil
	})
	if e != nil {
		log.Errorf("Fail to write redis,error:%v", err)
		return err
	}

	//comment
	commentLikeKeys, err := model.RedisDB.Keys("commentlike:*").Result() //会阻塞？
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	commentLikeValues, err := model.RedisDB.MGet(commentLikeKeys...).Result()
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	l1 = len(commentLikeKeys)

	commentLikesKeys, err := model.RedisDB.Keys("commentlikes:*").Result() //会阻塞？
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	commentLikesValues, err := model.RedisDB.MGet(commentLikeKeys...).Result()
	if err != nil {
		log.Errorf("Fail to read redis,error:%v", err)
		return err
	}
	l2 = len(commentLikeKeys)

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < l1; i++ {
			if commentLikeValues[i] == LIKE {
				params := strings.Split(commentLikeKeys[i], ":")
				comment, _ := strconv.Atoi(params[2])
				tx.Model(&model.CommentLikeShip{}).Create(&model.CommentLikeShip{
					User:         params[1],
					LikedComment: comment,
				})
			} else if commentLikeValues[i] == NEEDUNLIKE {
				params := strings.Split(commentLikeKeys[i], ":")
				tx.Model(&model.CommentLikeShip{}).Where("user=? AND liked_comment=?", params[1], params[2]).Delete(&model.CommentLikeShip{})
			}
		}

		for i := 0; i < l2; i++ {
			params := strings.Split(commentLikesKeys[i], ":")
			tx.Model(&model.Comment{}).Where("id=?", params[1]).Update("likes", commentLikesValues[i])
		}

		return nil
	})
	if err != nil {
		log.Errorf("Fail to finish postgres transaction,error:%v", err)
		return err
	}
	//redis里记录删了
	_, e = model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {

		for _, key := range commentLikeKeys {
			model.RedisDB.Del(key)
		}
		for _, key := range commentLikesKeys {
			model.RedisDB.Del(key)
		}
		return nil
	})
	if e != nil {
		log.Errorf("Fail to write redis,error:%v", err)
		return err
	}

	return nil
}
