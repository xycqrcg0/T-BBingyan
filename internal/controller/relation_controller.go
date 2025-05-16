package controller

import (
	"BBingyan/internal/config"
	"BBingyan/internal/controller/param"
	"BBingyan/internal/global"
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

func FollowUser(c echo.Context) error {
	user := c.Get("identification").(string)
	followed := c.Param("email")

	//先确定email的合法性
	emailKey := fmt.Sprintf("email:%s", followed)
	if v, err := model.RedisDB.Get(emailKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			if _, er := model.GetUserByEmail(followed); er != nil {
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

	err := model.FollowUser(user, followed)
	if err != nil {
		if errors.Is(err, global.ErrFollowExisted) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    err.Error(),
			})
		}
		log.Errorf("Fail to write in postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server",
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
	})
}

func UnFollowUser(c echo.Context) error {
	user := c.Get("identification").(string)
	unfollowed := c.Param("email")

	//先确定email的合法性
	emailKey := fmt.Sprintf("email:%s", unfollowed)
	if v, err := model.RedisDB.Get(emailKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			if _, er := model.GetUserByEmail(unfollowed); er != nil {
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

	err := model.UnfollowUser(user, unfollowed)
	if err != nil {
		if errors.Is(err, global.ErrFollowNonexistent) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    err.Error(),
			})
		}
		log.Errorf("Fail to write in postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server",
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
	})
}

// GetFollows 查询指定email用户的关注列表
func GetFollows(c echo.Context) error {
	user := c.Param("email")
	emailKey := fmt.Sprintf("email:%s", user)
	if v, err := model.RedisDB.Get(emailKey).Result(); err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("Fail to read redis,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal Server",
			})
		} else {
			if _, er := model.GetUserByEmail(user); er != nil {
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

	pageString := c.QueryParam("page")
	pageSizeString := c.QueryParam("page-size")
	page, _ := strconv.Atoi(pageString)
	pageSize, _ := strconv.Atoi(pageSizeString)
	if page <= 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = config.Config.Curd.PageSize
	}

	follows, err := model.GetAllFollows(user, page, pageSize)
	if err != nil {
		log.Errorf("Fail to read postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server",
		})
	}

	followsReq := make([]param.FollowUser, 0)
	for _, f := range follows {
		followsReq = append(followsReq, param.FollowUser{
			Email:     f.FollowedInfo.Email,
			Name:      f.FollowedInfo.Name,
			Signature: f.FollowedInfo.Signature,
			Likes:     f.FollowedInfo.Likes,
			Follows:   f.FollowedInfo.Follows,
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data: &param.FollowsResponse{
			Page:     page,
			PageSize: pageSize,
			Follows:  followsReq,
		},
	})
}

// GetFans 查询自己的关注者列表
func GetFans(c echo.Context) error {
	user := c.Get("identification").(string)
	pageString := c.QueryParam("page")
	pageSizeString := c.QueryParam("page-size")
	page, _ := strconv.Atoi(pageString)
	pageSize, _ := strconv.Atoi(pageSizeString)
	if page <= 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = config.Config.Curd.PageSize
	}

	fans, err := model.GetAllFans(user, page, pageSize)
	if err != nil {
		log.Errorf("Fail to read postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server",
		})
	}

	followsReq := make([]param.FollowUser, 0)
	for _, f := range fans {
		followsReq = append(followsReq, param.FollowUser{
			Email:     f.Info.Email,
			Name:      f.Info.Name,
			Signature: f.Info.Signature,
			Likes:     f.Info.Likes,
			Follows:   f.Info.Follows,
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data: &param.FollowsResponse{
			Page:     page,
			PageSize: pageSize,
			Follows:  followsReq,
		},
	})
}
