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

func AddPost(c echo.Context) error {
	user := c.Get("identification").(string)
	var data param.PostRequest
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	j := false
	for _, tag := range config.Config.Curd.Tags {
		if tag == data.Tag {
			j = true
			break
		}
	}
	if !j {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Nonexistent tag",
		})
	}

	if data.Title == "" || data.Content == "" {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid input",
		})
	}

	newPost := &model.Post{
		Author:  user,
		Title:   data.Title,
		Tag:     data.Tag,
		Content: data.Content,
	}

	err := model.AddPost(newPost)
	if err != nil {
		log.Errorf("Fail to write in postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server",
		})
	}

	return c.JSON(http.StatusCreated, &param.Response{
		Status: true,
		Msg:    "Create post Successfully",
	})
}

func DeletePost(c echo.Context) error {
	user := c.Get("identification").(string)
	ids := c.Param("id")
	id, _ := strconv.Atoi(ids)

	err := model.DeletePostById(user, id)
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
				Msg:    "Internal server",
			})
		}
	}
	if _, err := model.DeleteCommentsByPost(id); err != nil {
		log.Errorf("Fail to delete comments by pid,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "Delete post successfully",
	})
}

func GetPostByEmail(c echo.Context) error {
	email := c.Param("email")
	pageString := c.QueryParam("page")
	pageSizeString := c.QueryParam("page-size")
	page, _ := strconv.Atoi(pageString)
	pageSize, _ := strconv.Atoi(pageSizeString)
	if page < 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = config.Config.Curd.PageSize
	}

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

	posts, err := model.GetPostsByEmail(email, page, pageSize)
	if err != nil {
		log.Errorf("Fail to read postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal Server",
		})
	}
	postsResponse := make([]param.PostResponse, 0)
	for _, post := range posts {
		//点赞信息由于数据库里不是最新的，要从redis里再获取一遍？，这里就不错误处理了（懒）
		k1 := fmt.Sprintf("postlikes:%d", post.ID)
		k2 := fmt.Sprintf("userlikes:%s", post.User.Email)
		postlikes, e1 := model.RedisDB.Get(k1).Result()
		if e1 == nil {
			l, _ := strconv.Atoi(postlikes)
			post.Likes = l
		}
		userlikes, e2 := model.RedisDB.Get(k2).Result()
		if e2 == nil {
			l, _ := strconv.Atoi(userlikes)
			post.User.Likes = l
		}

		postsResponse = append(postsResponse, param.PostResponse{
			ID:        post.ID,
			Title:     post.Title,
			Tag:       post.Tag,
			Content:   post.Content,
			Likes:     post.Likes,
			Replies:   post.Replies,
			CreatedAt: post.CreatedAt,
			User: param.UserLessInfoResponse{
				Email:     post.User.Email,
				Name:      post.User.Name,
				Signature: post.User.Signature,
				Likes:     post.User.Likes,
				Follows:   post.User.Follows,
			},
		},
		)
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data:   postsResponse,
	})
}

func GetPostByTag(c echo.Context) error {
	tag := c.QueryParam("tag")
	ok := false
	for _, t := range config.Config.Curd.Tags {
		if t == tag {
			ok = true
			break
		}
	}
	if !ok {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "nonexistent tag",
		})
	}
	pageString := c.QueryParam("page")
	pageSizeString := c.QueryParam("page-size")
	page, _ := strconv.Atoi(pageString)
	pageSize, _ := strconv.Atoi(pageSizeString)
	if page < 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = config.Config.Curd.PageSize
	}

	ty := c.QueryParam("type")
	var err error
	var posts []model.Post
	switch ty {
	case param.TIMEDESC:
		posts, err = model.GetPostsByTagTime(tag, page, pageSize, true)
	case param.TIME:
		posts, err = model.GetPostsByTagTime(tag, page, pageSize, false)
	case param.REPLYDESC:
		posts, err = model.GetPostsByTagReplies(tag, page, pageSize, true)
	case param.REPLY:
		posts, err = model.GetPostsByTagReplies(tag, page, pageSize, false)
	default:
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Nonexistent type",
		})
	}
	if err != nil {
		log.Errorf("Fail to read from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal Server",
		})
	}

	postsResponse := make([]param.PostResponse, 0)
	for _, post := range posts {
		//点赞信息由于数据库里不是最新的，要从redis里再获取一遍？，这里就不错误处理了（懒）
		k1 := fmt.Sprintf("postlikes:%d", post.ID)
		k2 := fmt.Sprintf("userlikes:%s", post.User.Email)
		postlikes, e1 := model.RedisDB.Get(k1).Result()
		if e1 == nil {
			l, _ := strconv.Atoi(postlikes)
			post.Likes = l
		}
		userlikes, e2 := model.RedisDB.Get(k2).Result()
		if e2 == nil {
			l, _ := strconv.Atoi(userlikes)
			post.User.Likes = l
		}

		postsResponse = append(postsResponse, param.PostResponse{
			ID:        post.ID,
			Title:     post.Title,
			Tag:       post.Tag,
			Content:   post.Content,
			Likes:     post.Likes,
			Replies:   post.Replies,
			CreatedAt: post.CreatedAt,
			User: param.UserLessInfoResponse{
				Email:     post.User.Email,
				Name:      post.User.Name,
				Signature: post.User.Signature,
				Likes:     post.User.Likes,
				Follows:   post.User.Follows,
			},
		},
		)
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data:   postsResponse,
	})
}
