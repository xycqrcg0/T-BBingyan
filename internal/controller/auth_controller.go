package controller

import (
	"BBingyan/internal/controller/param"
	"BBingyan/internal/log"
	"BBingyan/internal/model"
	"BBingyan/internal/util"
	"errors"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func RegisterForCode(c echo.Context) error {
	//在query给个email就行
	//要看email是否符合格式
	email := c.Request().URL.Query().Get("email")

	redisKey := "register:email:" + email

	//先查该email是否在redis里有记录
	codeKey := "code:" + email
	_, err0 := model.RedisDB.Get(codeKey).Result()
	if err0 == nil {
		//有记录
		return c.JSON(http.StatusForbidden, &param.Response{
			Status: false,
			Msg:    "Please Wait for minutes",
		})
	} else {
		if !errors.Is(err0, redis.Nil) {
			log.Errorf("Fail to read code from redis,error:%v", err0)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	}

	//查email是否已存在
	if _, err := model.GetUserByEmail(email); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("Fail to get user from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	} else {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "email is existed",
		})
	}

	//生成验证码并发送
	//存下
	code := util.GenerateCode()
	_, err := model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {
		//限制发送
		pipe.Set(codeKey, "", time.Minute*1)
		//记录code
		pipe.Set(redisKey, code, time.Minute*5)
		return nil
	})
	if err != nil {
		log.Errorf("Fail to write Redis: keys=[%s, %s], error=%v", codeKey, redisKey, err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	//if err := util.SendAuthCode(email, code); err != nil {
	//	log.Errorf("Fail to send email,error:%v", err0)
	//	return c.JSON(http.StatusInternalServerError, &param.Response{
	//		Status: false,
	//		Msg:    "Internal server error",
	//	})
	//}
	util.SendAuthCode(email, code)

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "Send code successfully",
	})
}

func Register(c echo.Context) error {
	data := &param.UserRequest{}

	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    err.Error(),
		})
	}

	//数据都先检查一遍
	if data.Email == "" || data.Code == "" || data.Password == "" || data.Name == "" {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	//检查code
	redisKey := "register:email:" + data.Email
	com, err := model.RedisDB.Get(redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			//邮箱对应的code无/过期
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Wrong auth code",
			})
		}
		log.Errorf("Fail to read from redis for code,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}
	if data.Code != com {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Wrong auth code",
		})
	}

	//能到这说明验证码过了，可以注册
	data.Password, err = util.HashPwd(data.Password)
	if err != nil {
		log.Warnf("Fail to hash password,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	newUser := &model.User{
		Email:    data.Email,
		Name:     data.Name,
		Password: data.Password,
	}

	if err := model.AddUser(newUser); err != nil {
		log.Errorf("Fail to insert a user into postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	token, err := util.GenerateJWT(newUser.Email, param.USER)
	if err != nil {
		log.Warnf("Fail to generate jwt,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	return c.JSON(http.StatusCreated, &param.Response{
		Status: true,
		Msg:    "register successfully",
		Data: &param.TokenResponse{
			Token: token,
		},
	})
}

func LoginForCode(c echo.Context) error {
	//在query给个email就行
	//要看email是否符合格式
	email := c.Request().URL.Query().Get("email")

	redisKey := "login:email:" + email

	//先查该email是否在redis里有记录
	codeKey := "code:" + email
	_, err0 := model.RedisDB.Get(codeKey).Result()
	if err0 == nil {
		return c.JSON(http.StatusForbidden, &param.Response{
			Status: false,
			Msg:    "Please wait for minute",
		})
	} else {
		if !errors.Is(err0, redis.Nil) {
			log.Errorf("Fail to read code from redis,error:%v", err0)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	}

	//查email是否已存在
	if _, err := model.GetUserByEmail(email); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "email is not existed",
			})
		} else {
			log.Errorf("Fail to get user from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	}

	//生成验证码并发送
	//存下
	code := util.GenerateCode()
	_, err := model.RedisDB.TxPipelined(func(pipe redis.Pipeliner) error {
		//限制发送
		pipe.Set(codeKey, "", time.Minute*1)
		//记录code
		pipe.Set(redisKey, code, time.Minute*5)
		return nil
	})
	if err != nil {
		log.Errorf("Fail towrite in redis,error:%v", err0)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	//if err := util.SendAuthCode(email, code); err != nil {
	//	log.Errorf("Fail to send email,error:%v", err0)
	//	return c.JSON(http.StatusInternalServerError, &param.Response{
	//		Status: false,
	//		Msg:    "Internal server error",
	//	})
	//}
	util.SendAuthCode(email, code)

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "Login successfully",
	})
}

func Login(c echo.Context) error {
	style := c.Param("style")
	if style != "v1" && style != "v2" {
		return c.JSON(http.StatusNotFound, &param.Response{
			Status: false,
			Msg:    "",
		})
	}
	//v1->code;v2->password

	data := &param.UserRequest{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid Request",
		})
	}

	if style == "v1" {
		//code
		redisKey := "login:email:" + data.Email
		com, err0 := model.RedisDB.Get(redisKey).Result()
		if err0 != nil {
			if errors.Is(err0, redis.Nil) {
				return c.JSON(http.StatusBadRequest, &param.Response{
					Status: false,
					Msg:    "Please wait for minute",
				})
			}
			log.Errorf("Fail to read code from redis,error:%v", err0)
			//邮箱对应的code无/过期
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
		if data.Code != com {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Wrong auth code",
			})
		}
	} else {
		//password
		user, err := model.GetUserByEmail(data.Email)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusForbidden, &param.Response{
					Status: false,
					Msg:    "Wrong email or password",
				})
			}
			log.Errorf("Fail to read user from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
		if err := util.ParsePwd(user.Password, data.Password); err != nil {
			return c.JSON(http.StatusForbidden, &param.Response{
				Status: false,
				Msg:    "Wrong email or password",
			})
		}
	}

	token, err := util.GenerateJWT(data.Email, param.USER)
	if err != nil {
		log.Warnf("Fail to generate jwt,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "Login successfully",
		Data: &param.TokenResponse{
			Token: token,
		},
	})
}
