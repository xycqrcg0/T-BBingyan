package controller

import (
	"BBingyan/internal/controller/param"
	"BBingyan/internal/global"
	"BBingyan/internal/log"
	"BBingyan/internal/model"
	"BBingyan/internal/util"
	"errors"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

//admin应该就不要什么发评论啊，点赞什么的功能了，就留一些get和删评什么的功能

func AdminLogin(c echo.Context) error {
	data := &param.AdminReq{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid Request",
		})
	}

	admin, err := model.HasAdmin(data.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusForbidden, &param.Response{
				Status: false,
				Msg:    "Wrong name or password",
			})
		}
		log.Errorf("Fail to read admin from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}
	if err := util.ParsePwd(admin.Password, data.Password); err != nil {
		return c.JSON(http.StatusForbidden, &param.Response{
			Status: false,
			Msg:    "Wrong name or password",
		})
	}

	token, err := util.GenerateJWT(data.Name, param.ADMIN)
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

func CreateAdmin(c echo.Context) error {
	name := c.Get("identification").(string)
	data := &param.AdminReq{}
	if err := c.Bind(data); err != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	pwd, err := util.HashPwd(data.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    err.Error(),
		})
	}
	newAdmin := &model.Admin{
		Name:       data.Name,
		Password:   pwd,
		AddedAdmin: name,
	}

	if err := model.CreateAdmin(newAdmin); err != nil {
		log.Errorf("Fail to add new admin,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    err.Error(),
		})
	}

	log.Infof("admin [%s] create a new admin [%s] at %v", name, data.Name, time.Now())
	return c.JSON(http.StatusCreated, &param.Response{
		Status: true,
		Msg:    "",
		Data:   newAdmin,
	})
}

func DeleteAdmin(c echo.Context) error {
	name := c.Get("identification").(string)
	target := c.Param("name")
	if _, err := model.HasAdmin(target); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Nonexistent admin",
			})
		} else {
			log.Errorf("Fail to get admin info,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    err.Error(),
			})
		}
	}

	if err := model.DeleteAdmin(target); err != nil {
		log.Errorf("Fail to delete admin,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    err.Error(),
		})
	}

	log.Infof("admin [%s] delete an admin account [%s] at %v", name, target, time.Now())
	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
	})
}

func CreateTag(c echo.Context) error {
	name := c.Get("identification").(string)
	data := &model.Tag{}
	if err := c.Bind(data); err != nil || data.Tag == "" {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	tag := &model.Tag{
		Tag:        data.Tag,
		AddedAdmin: name,
	}
	if err := model.CreateTag(tag); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Repeated tag",
			})
		}
	}

	log.Infof("admin [%s] create a tag [%s] at %v", name, data.Tag, time.Now())
	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
	})
}

func AdminDeletePost(c echo.Context) error {
	ids := c.Param("id")
	id, _ := strconv.Atoi(ids)

	err := model.DeletePostOnlyById(id)
	if err != nil {
		if errors.Is(err, global.ErrPostNone) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Invalid request",
			})
		} else {
			log.Errorf("Fail to write in postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    err.Error(),
			})
		}
	}
	if _, err := model.DeleteCommentsByPost(id); err != nil {
		log.Errorf("Fail to delete comments by pid,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    err.Error(),
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "Delete post successfully",
	})
}

func AdminDeleteComment(c echo.Context) error {
	idstr := c.Param("id")
	id, e := strconv.Atoi(idstr)
	if e != nil {
		return c.JSON(http.StatusNotFound, &param.Response{
			Status: false,
			Msg:    "",
		})
	}

	var count int64
	var err error

	comment, err := model.GetCommentById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Invalid request",
			})
		} else {
			log.Errorf("Fail to get comment from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    err.Error(),
			})
		}
	}

	if comment.Root == 0 {
		//评论链删完
		count, err = model.DeleteRepliesByRoot(comment.Root)
		if err != nil {
			log.Errorf("Fail to delete comment from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    err.Error(),
			})
		}
	} else {
		//父评论的回复数要变了
		if err := model.ChangeCommentReplies(comment.Root, false); err != nil {
			log.Errorf("Fail to change comment replies from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    err.Error(),
			})
		}
	}

	if err := model.DeleteComment(id); err != nil {
		log.Errorf("Fail to delete father comment from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    err.Error(),
		})
	}
	count++

	if err := model.ChangePostReplies(comment.Pid, false, int(count)); err != nil {
		log.Errorf("Fail to change post replies from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "Delete comment successfully",
	})
}
