package global

import (
	"errors"
	"github.com/sirupsen/logrus"
)

var (
	Errors *logrus.Logger

	ErrPostNone          = errors.New("none")
	ErrFollowExisted     = errors.New("has followed")
	ErrFollowNonexistent = errors.New("hasn't followed")
)
