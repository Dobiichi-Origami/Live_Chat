package controllers

import (
	"liveChat/db"
	"strconv"
	"time"
)

func GetToken(userId int64) (token string, err error) {
	token = strconv.FormatInt(userId, 10) + "_" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	return db.RedisSetAndCheckTimeoutToken(token, userId)
}

func GetUserIdByToken(token string) (userId int64, err error) {
	return db.RedisCheckAndResetToken(token)
}
