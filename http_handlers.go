package liveChat

import "github.com/gin-gonic/gin"

func GetHandlerByName(string) gin.HandlerFunc {
	return nil
}

const (
	accountGetParam  = "account"
	passwordGetParam = "password"
)

func Login(ctx gin.Context) {
	_, ok := ctx.Get(accountGetParam)
	if !ok {

	}
}
