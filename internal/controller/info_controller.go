package controller

import (
	"BBingyan/internal/controller/param"
	"BBingyan/internal/log"
	"BBingyan/internal/model"
	"BBingyan/internal/util"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func GetUSerInfo(c echo.Context) error {
	useremail := c.Get("identification").(string)
	email := c.Param("email")
	//先确定email的合法性
	emailKey := fmt.Sprintf("email:%s", email)
	if v, err := model.RedisDB.Get(emailKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			if _, er := model.GetUserByEmail(email); er != nil {
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

	user, err := model.GetUserByEmail(email)
	if err != nil {
		log.Errorf("Fail to read postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal Server",
		})
	}

	k1 := fmt.Sprintf("userlikes:%s", user.Email)
	userlikes, e1 := model.RedisDB.Get(k1).Result()
	if e1 == nil {
		l, _ := strconv.Atoi(userlikes)
		user.Likes = l
	}

	//点赞、关注关系
	isLiked := false
	isFollowed := false

	lk := fmt.Sprintf("userlike:%s:%s", useremail, email)
	ld, e2 := model.RedisDB.Get(lk).Result()
	if e2 == nil {
		if ld == HASLIKED || ld == LIKE {
			isLiked = true
		}
	} else if errors.Is(e2, redis.Nil) {
		bl, e := model.HasLikeUserShip(useremail, email)
		if e != nil {
			log.Errorf("Fail to read postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		}
		isLiked = bl
	}
	fd, e3 := model.HasFollowed(useremail, email)
	if e3 != nil {
		log.Errorf("Fail to read postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal Server",
		})
	}
	isFollowed = fd

	userResponse := param.UserMoreInfoResponse{
		Email:      user.Email,
		Name:       user.Name,
		Signature:  user.Signature,
		Likes:      user.Likes,
		Follows:    user.Follows,
		CreatedAt:  user.CreatedAt,
		IsLiked:    isLiked,
		IsFollowed: isFollowed,
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data:   userResponse,
	})
}

func UpdateInfo(c echo.Context) error {
	email := c.Get("email").(string)

	data := &param.UserUpdateRequest{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	if data.Name == "" && data.Password == "" && data.Signature == "" {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	user, err := model.GetUserByEmail(email)
	if err != nil {
		log.Errorf("Fail to read user from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}
	if data.Name != "" {
		user.Name = data.Name
	}
	if data.Signature != "" {
		user.Signature = data.Signature
	}
	if data.Password != "" {
		b, err := util.HashPwd(data.Password)
		if err != nil {
			log.Warnf("Fail to hash password")
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
		user.Password = b
	}
	if err := model.UpdateUser(user); err != nil {
		log.Errorf("Fail to update user in postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}
	return c.JSON(http.StatusOK, "")
}
