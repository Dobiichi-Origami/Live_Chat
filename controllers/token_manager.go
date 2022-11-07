package controllers

import "liveChat/containers"

var tokenManager containers.ConcurrentMap

func init() {
	tokenManager = containers.New()
}

func SetToken(token string, userId int64) {
	tokenManager.Set(token, userId)
}

func GetUserIdByToken(token string) int64 {
	if ret, ok := tokenManager.Get(token); ok {
		return ret.(int64)
	}
	return -1
}
