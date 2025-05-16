package controller

import (
	"BBingyan/internal/config"
	"BBingyan/internal/controller/param"
	"BBingyan/internal/log"
	"BBingyan/internal/model"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreateComment(c echo.Context) error {
	email := c.Get("identification").(string)
	var commentReq param.CommentReq
	if err := c.Bind(&commentReq); err != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}
	if commentReq.Content == "" {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	//判断pid合理性
	if _, err := model.GetPostById(commentReq.Pid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Nonexistent post",
			})
		} else {
			log.Errorf("Fail to get post from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	}

	newComment := model.Comment{
		Uid:     email,
		Pid:     commentReq.Pid,
		Root:    commentReq.Root,
		Parent:  commentReq.Parent,
		Content: commentReq.Content,
		Replies: 0,
		Likes:   0,
	}

	if newComment.Root == 0 {
		//这个是父评论（一级）,parent直接置0
		newComment.Parent = 0
		if err := model.CreateComment(&newComment); err != nil {
			log.Errorf("Fail to write into postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	} else {
		//那么就是reply了
		//先查root合理性(父评论得是当前文章下的)
		if _, err := model.GetCommentByPid(newComment.Root, newComment.Pid); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusBadRequest, &param.Response{
					Status: false,
					Msg:    "Nonexistent root",
				})
			} else {
				log.Errorf("Fail to get comment from postgres,error:%v", err)
				return c.JSON(http.StatusInternalServerError, &param.Response{
					Status: false,
					Msg:    "Internal server error",
				})
			}
		}
		if commentReq.Parent != 0 {
			//还要查parent...(回复评论得是root评论下的)
			if _, err := model.GetReplyById(newComment.Parent, newComment.Root); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return c.JSON(http.StatusBadRequest, &param.Response{
						Status: false,
						Msg:    "Nonexistent parent",
					})
				} else {
					log.Errorf("Fail to get comment from postgres,error:%v", err)
					return c.JSON(http.StatusInternalServerError, &param.Response{
						Status: false,
						Msg:    "Internal server error",
					})
				}
			}
		}

		//可以写入了,这里有操作不同步的风险;到时候还要通知被回复用户,,,
		if err := model.CreateComment(&newComment); err != nil {
			log.Errorf("Fail to write into postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
		if err := model.ChangeCommentReplies(newComment.Root, true); err != nil {
			log.Errorf("Fail to write into postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	}

	if err := model.ChangePostReplies(newComment.Pid, true, 1); err != nil {
		log.Errorf("Fail to write into postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	return c.JSON(http.StatusCreated, &param.Response{
		Status: true,
		Msg:    "create comment successfully",
	})
}

func DeleteComment(c echo.Context) error {
	email := c.Get("identification").(string)
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

	//这里其实已经检查过email和id的对应关系了
	comment, err := model.GetCommentByUid(id, email)
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
				Msg:    "Internal server error",
			})
		}
	}

	if comment.Root == 0 {
		//这是要把这一条评论链删完啊，不然一级父评论没了，子评论怎么处理
		//考虑要不要一个个发消息通知？(好麻烦这样)
		count, err = model.DeleteRepliesByRoot(comment.Root)
		if err != nil {
			log.Errorf("Fail to delete comment from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	} else {
		//父评论的回复数要变了
		if err := model.ChangeCommentReplies(comment.Root, false); err != nil {
			log.Errorf("Fail to change comment replies from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	}

	if err := model.DeleteComment(id); err != nil {
		log.Errorf("Fail to delete father comment from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
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

func GetComments(c echo.Context) error {
	pidStr := c.QueryParam("pid")
	pageStr := c.QueryParam("page")
	pageSizeStr := c.QueryParam("page-size")
	ty := c.QueryParam("type")
	pid, e1 := strconv.Atoi(pidStr)
	var e2 error
	var e3 error
	pageSize := config.Config.Curd.PageSize
	page := 0
	if pageStr != "" {
		page, e2 = strconv.Atoi(pageStr)
		if page <= 0 {
			page = 0
		}
	}
	if pageSizeStr != "" {
		pageSize, e3 = strconv.Atoi(pageSizeStr)
		if pageSize <= 0 {
			pageSize = config.Config.Curd.PageSize
		}
	}
	if e1 != nil || e2 != nil || e3 != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}

	rawComments := make([]model.Comment, 0)
	var err error
	switch ty {
	case param.TIMEDESC:
		rawComments, err = model.GetCommentsByPost(pid, page, pageSize, true)
	case param.REPLYDESC:
		rawComments, err = model.GetCommentsByPost(pid, page, pageSize, false)
	default:
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}
	if err != nil {
		log.Errorf("Fail to get comments from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	commentRes := make([]param.CommentRes, 0)
	for _, comment := range rawComments {
		k1 := fmt.Sprintf("commentlikes:%d", comment.ID)
		k2 := fmt.Sprintf("userlikes:%s", comment.User.Email)
		commentlikes, e1 := model.RedisDB.Get(k1).Result()
		if e1 == nil {
			l, _ := strconv.Atoi(commentlikes)
			comment.Likes = l
		}
		userlikes, e2 := model.RedisDB.Get(k2).Result()
		if e2 == nil {
			l, _ := strconv.Atoi(userlikes)
			comment.User.Likes = l
		}
		commentRes = append(commentRes, param.CommentRes{
			Id:        comment.ID,
			Uid:       comment.Uid,
			Pid:       comment.Pid,
			Root:      comment.Root,
			Parent:    comment.Parent,
			Content:   comment.Content,
			Replies:   comment.Replies,
			Likes:     comment.Likes,
			CreatedAt: comment.CreatedAt,
			User: param.UserLessInfoResponse{
				Email:     comment.User.Email,
				Name:      comment.User.Name,
				Signature: comment.User.Signature,
				Likes:     comment.User.Likes,
				Follows:   comment.User.Follows,
			},
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data:   commentRes,
	})
}

func GetReplies(c echo.Context) error {
	pidStr := c.QueryParam("pid")
	rootStr := c.QueryParam("root")
	pageStr := c.QueryParam("page")
	pageSizeStr := c.QueryParam("page-size")
	ty := c.QueryParam("type")
	page := 0
	pageSize := config.Config.Curd.PageSize
	var e1 error
	var e2 error
	if pageStr != "" {
		page, e1 = strconv.Atoi(pageStr)
		if page <= 0 {
			page = 0
		}
	}
	if pageSizeStr != "" {
		pageSize, e2 = strconv.Atoi(pageSizeStr)
		if pageSize <= 0 {
			pageSize = config.Config.Curd.PageSize
		}
	}
	root, e3 := strconv.Atoi(rootStr)
	pid, e4 := strconv.Atoi(pidStr)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}
	//检查root
	if _, err := model.GetCommentByPid(root, pid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusBadRequest, &param.Response{
				Status: false,
				Msg:    "Invalid request",
			})
		} else {
			log.Errorf("Fail to get comment from postgres,error:%v", err)
			return c.JSON(http.StatusInternalServerError, &param.Response{
				Status: false,
				Msg:    "Internal server error",
			})
		}
	}

	rawReplies := make([]model.Comment, 0)
	var err error
	switch ty {
	case param.TIMEDESC:
		rawReplies, err = model.GetRepliesByRoot(pid, page, pageSize, model.TIMEDESC)
	case param.REPLYDESC:
		rawReplies, err = model.GetRepliesByRoot(pid, page, pageSize, param.REPLYDESC)
	default:
		return c.JSON(http.StatusBadRequest, &param.Response{
			Status: false,
			Msg:    "Invalid request",
		})
	}
	if err != nil {
		log.Errorf("Fail to get comments from postgres,error:%v", err)
		return c.JSON(http.StatusInternalServerError, &param.Response{
			Status: false,
			Msg:    "Internal server error",
		})
	}

	repliesRes := make([]param.CommentRes, 0)
	for _, comment := range rawReplies {
		k1 := fmt.Sprintf("commentlikes:%d", comment.ID)
		k2 := fmt.Sprintf("userlikes:%s", comment.User.Email)
		commentlikes, e1 := model.RedisDB.Get(k1).Result()
		if e1 == nil {
			l, _ := strconv.Atoi(commentlikes)
			comment.Likes = l
		}
		userlikes, e2 := model.RedisDB.Get(k2).Result()
		if e2 == nil {
			l, _ := strconv.Atoi(userlikes)
			comment.User.Likes = l
		}
		repliesRes = append(repliesRes, param.CommentRes{
			Id:        comment.ID,
			Uid:       comment.Uid,
			Pid:       comment.Pid,
			Root:      comment.Root,
			Parent:    comment.Parent,
			Content:   comment.Content,
			Replies:   comment.Replies,
			Likes:     comment.Likes,
			CreatedAt: comment.CreatedAt,
			User: param.UserLessInfoResponse{
				Email:     comment.User.Email,
				Name:      comment.User.Name,
				Signature: comment.User.Signature,
				Likes:     comment.User.Likes,
				Follows:   comment.User.Follows,
			},
		})
	}

	return c.JSON(http.StatusOK, &param.Response{
		Status: true,
		Msg:    "",
		Data:   repliesRes,
	})
}
