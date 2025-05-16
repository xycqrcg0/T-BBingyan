package controller

import (
	"BBingyan/internal/controller/param"
	"BBingyan/internal/log"
	"BBingyan/internal/model"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

//这部分代码都是复制粘贴哩~

var (
	HASLIKED   = "2"
	LIKE       = "1"
	UNLIKE     = "0"
	NEEDUNLIKE = "-1"
)

func LikeUser(c echo.Context) error {
	email := c.Get("identification").(string)
	LikedEmail := c.Param("email")

	//先确定email的合法性
	emailKey := fmt.Sprintf("email:%s", LikedEmail)
	if v, err := model.RedisDB.Get(emailKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			if _, er := model.GetUserByEmail(LikedEmail); er != nil {
				if errors.Is(er, gorm.ErrRecordNotFound) {
					if _, e := model.RedisDB.Set(emailKey, param.INVALID, time.Minute*5).Result(); e != nil {
						log.Errorf("Fail to write in redis,error:%v", err)
						return c.JSON(http.StatusInternalServerError, &param.Response{
							Status: false,
							Msg:    "Internal Server",
						})
					}
					return c.JSON(http.StatusBadRequest, &param.Response{
						Status: false,
						Msg:    "nonexistent email",
					})
				} else {
					log.Errorf("Fail to read postgres,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			} else {
				if _, e := model.RedisDB.Set(emailKey, param.VALID, time.Minute*5).Result(); e != nil {
					log.Errorf("Fail to write in redis,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			}
		}
	} else if v == param.INVALID {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "nonexistent email",
		})
	}

	likeKey := fmt.Sprintf("userlike:%s:%s", email, LikedEmail)
	likesKey := fmt.Sprintf("userlikes:%s", LikedEmail)

	//再获取点赞数/创建该kv对
	if _, err := model.RedisDB.Get(likesKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			//读数据库
			if l, err := model.GetUserLikes(LikedEmail); err != nil {
				log.Errorf("Fail to read postgres,error:%v", err)
				return c.JSON(http.StatusInternalServerError, &param.Response{
					Status: false,
					Msg:    "Internal Server",
				})
			} else {
				if _, er := model.RedisDB.Set(likesKey, l, time.Hour*5).Result(); er != nil {
					log.Errorf("Fail to write in redis,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			}
		}
	}

	//j用来判断之后要怎么记录kv对
	j, err := model.RedisDB.Get(likeKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			//查数据库
			ok, er := model.HasLikeUserShip(email, LikedEmail)
			if er != nil {
				log.Errorf("Fail to read postgres,error:%v", er)
				return c.JSON(http.StatusInternalServerError, &param.Response{
					Status: false,
					Msg:    "Internal Server",
				})
			}
			if !ok {
				//数据库里没有该条记录 -> 未写入的未点赞状态
				j = UNLIKE
			} else {
				//数据库里已经有该条记录 -> 已写入的已点赞状态
				j = HASLIKED
			}
		}
	}

	likes := "默认"
	_, er := model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {
		switch j {
		case HASLIKED:
			pipe.Set(likeKey, NEEDUNLIKE, time.Hour*5)
			pipe.Decr(likesKey)
			break
		case NEEDUNLIKE:
			pipe.Set(likeKey, HASLIKED, time.Hour*5)
			pipe.Incr(likesKey)
			break
		case UNLIKE:
			pipe.Set(likeKey, LIKE, time.Hour*5)
			pipe.Incr(likesKey)
			break
		case LIKE:
			pipe.Set(likeKey, UNLIKE, time.Hour*5)
			pipe.Decr(likesKey)
			break
		}
		return nil
	})
	if er != nil {
		log.Errorf("Fail to finish redis transaction,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal Server",
		})
	}
	likes, _ = model.RedisDB.Get(likesKey).Result()

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data: &param.LikeResponse{
			Likes: likes,
		},
	})
}

func LikePost(c echo.Context) error {
	email := c.Get("identification").(string)
	ID := c.Param("id")
	id, _ := strconv.Atoi(ID)

	//先确定post的合法性
	postKey := fmt.Sprintf("post:%d", id)
	if v, err := model.RedisDB.Get(postKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			if _, er := model.GetPostById(id); er != nil {
				if errors.Is(er, gorm.ErrRecordNotFound) {
					if _, e := model.RedisDB.Set(postKey, param.INVALID, time.Minute*5).Result(); e != nil {
						log.Errorf("Fail to write in redis,error:%v", err)
						return c.JSON(http.StatusInternalServerError, &param.Response{
							Status: false,
							Msg:    "Internal Server",
						})
					}
					return c.JSON(http.StatusBadRequest, &param.Response{
						Status: false,
						Msg:    "nonexistent post",
					})
				} else {
					log.Errorf("Fail to read postgres,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			} else {
				if _, e := model.RedisDB.Set(postKey, param.VALID, time.Minute*5).Result(); e != nil {
					log.Errorf("Fail to write in redis,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			}
		}
	} else if v == param.INVALID {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "nonexistent email",
		})
	}

	likeKey := fmt.Sprintf("postlike:%s:%d", email, id)
	likesKey := fmt.Sprintf("postlikes:%d", id)

	//再获取点赞数/创建该kv对
	if _, err := model.RedisDB.Get(likesKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			//读数据库
			if l, err := model.GetPostLikes(id); err != nil {
				log.Errorf("Fail to read postgres,error:%v", err)
				return c.JSON(http.StatusInternalServerError, &param.Response{
					Status: false,
					Msg:    "Internal Server",
				})
			} else {
				if _, er := model.RedisDB.Set(likesKey, l, time.Hour*5).Result(); er != nil {
					log.Errorf("Fail to write in redis,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			}
		}
	}

	//j用来判断之后要怎么记录kv对
	j, err := model.RedisDB.Get(likeKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			//查数据库
			ok, er := model.HasLikePostShip(email, id)
			if er != nil {
				log.Errorf("Fail to read postgres,error:%v", er)
				return c.JSON(http.StatusInternalServerError, &param.Response{
					Status: false,
					Msg:    "Internal Server",
				})
			}
			if !ok {
				//数据库里没有该条记录 -> 未写入的未点赞状态
				j = UNLIKE
			} else {
				//数据库里已经有该条记录 -> 已写入的已点赞状态
				j = HASLIKED
			}
		}
	}

	var likes string
	_, er := model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {
		switch j {
		case HASLIKED:
			pipe.Set(likeKey, NEEDUNLIKE, time.Hour*5)
			pipe.Decr(likesKey)
			break
		case NEEDUNLIKE:
			pipe.Set(likeKey, HASLIKED, time.Hour*5)
			pipe.Incr(likesKey)
			break
		case UNLIKE:
			pipe.Set(likeKey, LIKE, time.Hour*5)
			pipe.Incr(likesKey)
			break
		case LIKE:
			pipe.Set(likeKey, UNLIKE, time.Hour*5)
			pipe.Decr(likesKey)
			break
		}
		return nil
	})
	if er != nil {
		log.Errorf("Fail to finish redis transaction,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal Server",
		})
	}
	likes, _ = model.RedisDB.Get(likesKey).Result()

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data: &param.LikeResponse{
			Likes: likes,
		},
	})
}

func LikeComment(c echo.Context) error {
	email := c.Get("identification").(string)
	ID := c.Param("id")
	id, _ := strconv.Atoi(ID)


	//先确定post的合法性
	commentKey := fmt.Sprintf("comment:%d", id)
	if v, err := model.RedisDB.Get(commentKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			if _, er := model.GetCommentById(id); er != nil {
				if errors.Is(er, gorm.ErrRecordNotFound) {
					if _, e := model.RedisDB.Set(commentKey, param.INVALID, time.Minute*5).Result(); e != nil {
						log.Errorf("Fail to write in redis,error:%v", err)
						return c.JSON(http.StatusInternalServerError, &param.Response{
							Status: false,
							Msg:    "Internal Server",
						})
					}
					return c.JSON(http.StatusBadRequest, &param.Response{
						Status: false,
						Msg:    "nonexistent comment",
					})
				} else {
					log.Errorf("Fail to read postgres,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			} else {
				if _, e := model.RedisDB.Set(commentKey, param.VALID, time.Minute*5).Result(); e != nil {
					log.Errorf("Fail to write in redis,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			}
		}
	} else if v == param.INVALID {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "nonexistent email",
		})
	}

	likeKey := fmt.Sprintf("commentlike:%s:%d", email, id)
	likesKey := fmt.Sprintf("commmentlikes:%d", id)

	//再获取点赞数/创建该kv对
	if _, err := model.RedisDB.Get(likesKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			//读数据库
			if l, err := model.GetCommentLikes(id); err != nil {
				log.Errorf("Fail to read postgres,error:%v", err)
				return c.JSON(http.StatusInternalServerError, &param.Response{
					Status: false,
					Msg:    "Internal Server",
				})
			} else {
				if _, er := model.RedisDB.Set(likesKey, l, time.Hour*5).Result(); er != nil {
					log.Errorf("Fail to write in redis,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal Server",
					})
				}
			}
		}
	}

	//j用来判断之后要怎么记录kv对
	j, err := model.RedisDB.Get(likeKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			//查数据库
			ok, er := model.HasLikeCommentShip(email, id)
			if er != nil {
				log.Errorf("Fail to read postgres,error:%v", er)
				return c.JSON(http.StatusInternalServerError, &param.Response{
					Status: false,
					Msg:    "Internal Server",
				})
			}
			if !ok {
				//数据库里没有该条记录 -> 未写入的未点赞状态
				j = UNLIKE
			} else {
				//数据库里已经有该条记录 -> 已写入的已点赞状态
				j = HASLIKED
			}
		}
	}

	var likes string
	_, er := model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {
		switch j {
		case HASLIKED:
			pipe.Set(likeKey, NEEDUNLIKE, time.Hour*5)
			pipe.Decr(likesKey)
			break
		case NEEDUNLIKE:
			pipe.Set(likeKey, HASLIKED, time.Hour*5)
			pipe.Incr(likesKey)
			break
		case UNLIKE:
			pipe.Set(likeKey, LIKE, time.Hour*5)
			pipe.Incr(likesKey)
			break
		case LIKE:
			pipe.Set(likeKey, UNLIKE, time.Hour*5)
			pipe.Decr(likesKey)
			break
		}
		return nil
	})
	if er != nil {
		log.Errorf("Fail to finish redis transaction,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal Server",
		})
	}
	likes, _ = model.RedisDB.Get(likesKey).Result()

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data: &param.LikeResponse{
			Likes: likes,
		},
	})
}
