package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"liveChat/controllers"
	"liveChat/db"
	"liveChat/entities"
	"liveChat/log"
	"strconv"
	"time"
)

const (
	accountGetParam       = "account"
	passwordGetParam      = "password"
	emailGetParam         = "email"
	userIdParam           = "id"
	usernameParam         = "username"
	userAvatarParam       = "avatar"
	userIntroductionParam = "introduction"
	friendIdParam         = "friendId"

	groupIdParam           = "groupId"
	groupNameParam         = "groupName"
	groupIntroductionParam = "groupIntroduction"
	groupAvatarParam       = "groupAvatar"

	notificationSeqParam = "notificationSeq"

	tokenHeaderParam = "x-custom-token"
)

const (
	accountKey          = "account"
	passwordKey         = "password"
	emailKey            = "email"
	userIdKey           = "userId"
	tokenKey            = "token"
	usernameKey         = "username"
	userAvatarKey       = "avatar"
	userIntroductionKey = "introduction"
	friendIdKey         = "friendId"

	groupIdKey           = "groupId"
	groupNameKey         = "groupName"
	groupIntroductionKey = "groupIntroduction"
	groupAvatarKey       = "groupAvatar"

	groupMemberIsOwnerKey = "groupMemberIsOwner"
	groupMemberIsAuthKey  = "groupMemberIsAuth"
	groupMemberIsInKey    = "groupMemberIsIn"

	chatIdKey = "chatId"

	notificationSeqKey      = "notificationSeq"
	notificationReceiverKey = "notificationReceiver"

	isSameUserKey      = "isSameUser"
	userIdFromTokenKey = "userIdFromToken"
)

const (
	contentTypeJson = "application/json"
)

const (
	Success = 200

	UserParamTypeIllegal         = 419
	LackOfParameter              = 420
	UserNotFound                 = 421
	RegisterInfoValidateFailed   = 422
	TokenInvalid                 = 423
	IllegalRequestFromMismatched = 424
	GroupNotFound                = 425
	GroupOpNoAuth                = 426
	IllegalRequest               = 427
	InternalError                = 500
)

type ResponseHeader struct {
	Status int32 `json:"status"`
}

type SuccessBody struct {
	ResponseHeader
}

type FailBody struct {
	ResponseHeader
	Reason string `json:"reason"`
}

type RegisterOrLoginBody struct {
	ResponseHeader
	Token  string `json:"token"`
	UserId int64  `json:"userId"`
}

type UserInfoBody struct {
	ResponseHeader
	entities.UserInfo
}

type GroupInfoBody struct {
	ResponseHeader
	entities.GroupInfo
}

type FriendshipBody struct {
	ResponseHeader
	entities.Friendship
}

type createGroupForm struct {
	GroupName         string `json:"group_name"`
	GroupIntroduction string `json:"group_introduction"`
	GroupAvatar       string `json:"group_avatar"`
}

func postHandler(ctx *controllers.ProcessContext, retBuf []byte, err error) {
	ginCtx := ctx.Ctx.(*gin.Context)

	if err != nil {
		ginCtx.JSON(500, nil)
	} else {
		ginCtx.Data(Success, contentTypeJson, retBuf)
	}
}

func getAccountFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	account, retBuf, err := getParamFromURL(ctx, accountGetParam, "缺少登录账号", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[accountKey] = account
	}
	return
}

func getPasswordFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	password, retBuf, err := getParamFromURL(ctx, passwordGetParam, "缺少登录密码", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[passwordKey] = password
	}
	return
}

func getEmailFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	email, retBuf, err := getParamFromURL(ctx, emailGetParam, "缺少注册邮箱", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[emailKey] = email
	}
	return
}

func getUserIdFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	userId, retBuf, err := getInt64ParamFromURL(ctx, userIdParam, "缺少用户 Id", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[userIdKey] = userId
	}
	return
}

func getUsernameFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	username, retBuf, err := getParamFromURL(ctx, usernameParam, "缺少用户名", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[usernameKey] = username
	}
	return
}

func getUserAvatarFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	avatar, retBuf, err := getParamFromURL(ctx, userAvatarParam, "缺少用户名", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[userAvatarKey] = avatar
	}
	return
}

func getUserIntroductionFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	intro, retBuf, err := getParamFromURL(ctx, userIntroductionParam, "缺少用户名", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[userIntroductionKey] = intro
	}
	return
}

func getFriendIdFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	friendId, retBuf, err := getInt64ParamFromURL(ctx, friendIdParam, "缺少目标用户 id", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[friendIdKey] = friendId
	}
	return
}

func getGroupIdFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	groupId, retBuf, err := getInt64ParamFromURL(ctx, groupIdParam, "缺少目标群组 id", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[groupIdKey] = groupId
	}
	return
}

func getGroupNameFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	groupName, retBuf, err := getParamFromURL(ctx, groupNameParam, "缺少群名", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[groupNameKey] = groupName
	}
	return
}

func getGroupIntroductionFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	groupIntroduction, retBuf, err := getParamFromURL(ctx, groupIntroductionParam, "缺少群介绍", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[groupIntroductionKey] = groupIntroduction
	}
	return
}

func getGroupAvatarFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	groupAvatar, retBuf, err := getParamFromURL(ctx, groupAvatarParam, "缺少群头像", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[groupAvatarKey] = groupAvatar
	}
	return
}

func getNotificationSeqFromUrl(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	seq, retBuf, err := getParamFromURL(ctx, notificationSeqParam, "缺少需要确认的通知序号", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[notificationSeqKey] = seq
	}
	return
}

func getFriendIdAsNotificationReceiver(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	ctx.Param[notificationReceiverKey] = ctx.Param[friendIdKey]
	return
}

func getGroupIdAsNotificationReceiver(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	groupId, retBuf, err := getParamFromURL(ctx, groupIdParam, "缺少目标群组 id", LackOfParameter)
	if len(retBuf) == 0 && err == nil {
		ctx.Param[notificationReceiverKey] = groupId
	}
	return
}

func getTokenFromHeader(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	ginCtx := ctx.Ctx.(*gin.Context)
	token := ginCtx.GetHeader(tokenHeaderParam)
	if token == "" {
		retBuf, err = errorHandlerHook(LackOfParameter, fmt.Sprintf("Header 中缺少 %s 字段", tokenHeaderParam))
		return
	}

	ctx.Param[tokenKey] = token
	return
}

func getCreateGroupPostInBody(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	ginCtx := ctx.Ctx.(*gin.Context)

	createForm := &createGroupForm{}
	err = ginCtx.BindJSON(createForm)
	if err != nil {
		retBuf, err = errorHandlerHook(IllegalRequestFromMismatched, "Json 表单解析错误")
		return
	}

	ctx.Param[groupNameKey] = createForm.GroupName
	ctx.Param[groupIntroductionKey] = createForm.GroupIntroduction
	ctx.Param[groupAvatarKey] = createForm.GroupAvatar
	return
}

func registerUserInfoAndGetUserId(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		account  = ctx.Param[accountKey].(string)
		email    = ctx.Param[emailKey].(string)
		password = ctx.Param[passwordKey].(string)
	)

	userId, err := db.Register(nil, account, email, password)
	if err != nil {
		return
	}

	ctx.Param[userIdKey] = userId
	return
}

func getUserIdByAccountAndPassword(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		account  = ctx.Param[accountKey].(string)
		password = ctx.Param[passwordKey].(string)
	)

	userId, err := db.Login(nil, account, password)
	if err == db.MysqlErrorUserNotExist {
		retBuf, err = errorHandlerHook(UserNotFound, "请检查您的账号或密码是否正确")
		return
	} else if err != nil {
		return
	}

	ctx.Param[userIdKey] = userId
	return
}

func getTokenByUserId(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	userId := ctx.Param[userIdKey].(int64)
	token, err := controllers.GetToken(userId)
	if err != nil {
		return
	}

	ctx.Param[tokenKey] = token
	return
}

func validateToken(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	token := ctx.Param[tokenKey].(string)
	tokenUserId, err := controllers.GetUserIdByToken(token)
	if err != nil {
		return
	} else if tokenUserId == -1 {
		retBuf, err = errorHandlerHook(TokenInvalid, "token 无效")
		return
	}

	ctx.Param[userIdFromTokenKey] = tokenUserId
	return
}

func validateRegisterInfo(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		account  = ctx.Param[accountKey].(string)
		email    = ctx.Param[emailKey].(string)
		password = ctx.Param[passwordKey].(string)
	)

	if !validate(account, email, password) {
		retBuf, err = errorHandlerHook(RegisterInfoValidateFailed, "注册信息校验失败")
		return
	}
	return
}

func tellIsSameUserCompareTokenAndUserId(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userIdFromToken = ctx.Param[userIdFromTokenKey].(int64)
		userId          = ctx.Param[userIdKey].(int64)
	)

	ctx.Param[isSameUserKey] = userIdFromToken == userId
	return
}

func tellIsSameUserCompareTokenAndFriendId(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userIdFromToken = ctx.Param[userIdFromTokenKey].(int64)
		friendId        = ctx.Param[friendIdKey].(int64)
	)

	ctx.Param[isSameUserKey] = userIdFromToken == friendId
	return
}

func rejectRequestFromOther(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	if !ctx.Param[isSameUserKey].(bool) {
		retBuf, err = errorHandlerHook(IllegalRequestFromMismatched, "非法访问")
	}
	return
}

func rejectRequestFromOneSelf(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	if ctx.Param[isSameUserKey].(bool) {
		retBuf, err = errorHandlerHook(IllegalRequestFromMismatched, "非法操作")
	}
	return
}

func sendAddFriendNotificationToOther(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	noti := entities.NewNotification(ctx.Param[userIdFromTokenKey].(int64),
		ctx.Param[friendIdKey].(int64),
		entities.Add,
		entities.Friend,
		false,
		false)

	noti, err = db.AddAndReturnNotification(context.Background(), noti)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
		return
	}

	SendNotification(noti)
	return
}

func sendDeleteFriendNotificationToOther(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		selfId   = ctx.Param[userIdFromTokenKey].(int64)
		friendId = ctx.Param[friendIdKey].(int64)
	)

	flag, err := controllers.CheckAreUsersFriend(selfId, friendId, true)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
		return
	}

	if !flag {
		retBuf, err = errorHandlerHook(IllegalRequestFromMismatched, "非好友关系")
		return
	}

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		if err = db.DeleteFriendShip(mysqlTx, selfId, friendId); err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		noti := entities.NewNotification(selfId, friendId, entities.Delete, entities.Friend, true, true)
		noti.HandleUserId = selfId
		noti, err = db.AddAndReturnNotification(mongoTx, noti)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		if _, err = controllers.CheckAreUsersFriend(selfId, friendId, true); err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		SendNotification(noti)
		return nil
	})
	return
}

func approveFriendRequest(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId                        = ctx.Param[userIdFromTokenKey].(int64)
		seq                           = ctx.Param[notificationSeqKey].(uint64)
		noti   *entities.Notification = nil
		chatId                        = int64(0)
	)

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		noti, err = db.HandleNotification(mongoTx, userId, userId, seq, true)
		if err == db.MongoErrorNoNotification {
			retBuf, err = errorHandlerHook(IllegalRequestFromMismatched, "无相关通知")
			return err
		} else if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		chatId, err = db.AgreeFriendShip(mysqlTx, noti.SenderId, noti.ReceiverId)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}
		ctx.Param[chatIdKey] = chatId
		ctx.Param[friendIdKey] = noti.SenderId

		sendNoti := entities.NewNotification(userId, noti.SenderId, entities.Approve, entities.Friend, true, true)
		sendNoti, err = db.AddAndReturnNotification(mongoTx, sendNoti)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		if _, err = controllers.CheckAreUsersFriend(noti.SenderId, noti.ReceiverId, true); err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		SendNotification(sendNoti)
		return nil
	})
	return
}

func refuseFriendRequest(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId = ctx.Param[userIdFromTokenKey].(int64)
		seq    = ctx.Param[notificationSeqKey].(uint64)
	)

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		var noti *entities.Notification
		noti, err = db.HandleNotification(mongoTx, userId, userId, seq, false)
		if err == db.MongoErrorNoNotification {
			retBuf, err = errorHandlerHook(IllegalRequestFromMismatched, "无相关通知")
			return err
		} else if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		sendNoti := entities.NewNotification(userId, noti.ReceiverId, entities.Refuse, entities.Friend, true, false)
		sendNoti, err = db.AddAndReturnNotification(mongoTx, sendNoti)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		SendNotification(sendNoti)
		return nil
	})
	return
}

func updateUsername(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId   = ctx.Param[userIdFromTokenKey].(int64)
		username = ctx.Param[usernameKey].(string)
	)

	err = db.UpdateUserName(nil, userId, username)
	if err == db.MysqlErrorUserNotExist {
		retBuf, err = errorHandlerHook(UserNotFound, "目标用户不存在")
	}
	return
}

func updateUserAvatar(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId     = ctx.Param[userIdFromTokenKey].(int64)
		userAvatar = ctx.Param[userAvatarKey].(string)
	)

	err = db.UpdateUserAvatar(nil, userId, userAvatar)
	if err == db.MysqlErrorUserNotExist {
		retBuf, err = errorHandlerHook(UserNotFound, "目标用户不存在")
	}
	return
}

func updateUserIntroduction(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId       = ctx.Param[userIdFromTokenKey].(int64)
		introduction = ctx.Param[userIntroductionKey].(string)
	)

	err = db.UpdateUserIntroduction(nil, userId, introduction)
	if err == db.MysqlErrorUserNotExist {
		retBuf, err = errorHandlerHook(UserNotFound, "目标用户不存在")
	}
	return
}

func getGroupInfoBody(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId  = ctx.Param[userIdFromTokenKey].(int64)
		groupId = ctx.Param[groupIdKey].(int64)
	)

	flag, err := controllers.CheckIsUserInGroup(userId, groupId, false)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
		return
	}

	info, err := db.SearchGroupInfo(nil, groupId, flag)
	if err == db.MysqlErrorGroupNotExist {
		retBuf, err = errorHandlerHook(GroupNotFound, "群组不存在")
		return
	}

	err = (&GroupInfoBody{
		ResponseHeader: ResponseHeader{Success},
		GroupInfo:      *info,
	}).UnmarshalJSON(retBuf)
	return
}

func createGroup(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		groupName         = ctx.Param[groupNameKey].(string)
		groupIntroduction = ctx.Param[groupIntroductionKey].(string)
		groupAvatar       = ctx.Param[groupAvatarKey].(string)
		userId            = ctx.Param[userIdFromTokenKey].(int64)
	)

	groupId, err := db.AddGroupInfo(nil, userId, groupName, groupIntroduction, groupAvatar)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
		return
	}

	ctx.Param[groupIdKey] = groupId
	return
}

func checkGroupAuth(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId  = ctx.Param[userIdFromTokenKey].(int64)
		groupId = ctx.Param[groupIdKey].(int64)
	)

	info, err := db.SearchGroupInfo(nil, groupId, true)
	if err == db.MysqlErrorGroupNotExist {
		retBuf, err = errorHandlerHook(GroupNotFound, "群组不存在")
	} else if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
		return
	}

	if info.IsDeleted {
		retBuf, err = errorHandlerHook(GroupNotFound, "群组已解散")
		return
	}

	ctx.Param[groupMemberIsOwnerKey] = info.Owner == userId
	ctx.Param[groupMemberIsInKey] = false

	for _, member := range info.Members {
		if member.MemberId == userId {
			ctx.Param[groupMemberIsAuthKey] = member.IsAdministrator
			ctx.Param[groupMemberIsInKey] = true
			return
		}
	}

	ctx.Param[groupMemberIsAuthKey] = false
	return
}

func deleteGroup(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId  = ctx.Param[userIdFromTokenKey].(int64)
		groupId = ctx.Param[groupIdKey].(int64)
		isOwner = ctx.Param[groupMemberIsOwnerKey].(bool)
	)

	if !isOwner {
		retBuf, err = errorHandlerHook(GroupOpNoAuth, "无权限")
	}

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		err = db.DeleteGroupInfo(mysqlTx, groupId)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		_, err = controllers.CheckIsUserInGroup(userId, groupId, true)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		return nil
	})

	return
}

func updateGroupName(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		groupId   = ctx.Param[groupIdKey].(int64)
		groupName = ctx.Param[groupNameKey].(string)
		isAuth    = ctx.Param[groupMemberIsAuthKey].(bool)
	)

	if !isAuth {
		retBuf, err = errorHandlerHook(GroupOpNoAuth, "无权限")
		return
	}

	err = db.UpdateGroupName(nil, groupId, groupName)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
	}
	return
}

func updateGroupIntroduction(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		groupId           = ctx.Param[groupIdKey].(int64)
		groupIntroduction = ctx.Param[groupIntroductionKey].(string)
		isAuth            = ctx.Param[groupMemberIsAuthKey].(bool)
	)

	if !isAuth {
		retBuf, err = errorHandlerHook(GroupOpNoAuth, "无权限")
		return
	}

	err = db.UpdateGroupIntroduction(nil, groupId, groupIntroduction)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
	}
	return
}

func updateGroupAvatar(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		groupId     = ctx.Param[groupIdKey].(int64)
		groupAvatar = ctx.Param[groupAvatarKey].(string)
		isAuth      = ctx.Param[groupMemberIsAuthKey].(bool)
	)

	if !isAuth {
		retBuf, err = errorHandlerHook(GroupOpNoAuth, "无权限")
		return
	}

	err = db.UpdateGroupAvatar(nil, groupId, groupAvatar)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
	}
	return
}

func sendJoinGroupNotification(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		isIn    = ctx.Param[groupMemberIsInKey].(bool)
		userId  = ctx.Param[userIdFromTokenKey].(int64)
		groupId = ctx.Param[groupIdKey].(int64)
	)

	if isIn {
		retBuf, err = errorHandlerHook(IllegalRequest, "用户已在群组中")
		return
	}

	noti := entities.NewNotification(userId, groupId, entities.Add, entities.Group, false, false)
	noti, err = db.AddAndReturnNotification(context.Background(), noti)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
		return
	}

	SendNotification(noti)
	return
}

func approveJoinGroupRequest(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId                         = ctx.Param[userIdFromTokenKey].(int64)
		groupId                        = ctx.Param[notificationReceiverKey].(int64)
		seq                            = ctx.Param[notificationSeqKey].(uint64)
		isAuth                         = ctx.Param[groupMemberIsAuthKey].(bool)
		noti    *entities.Notification = nil
	)

	if !isAuth {
		retBuf, err = errorHandlerHook(GroupOpNoAuth, "无权限")
		return
	}

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		noti, err = db.HandleNotification(mongoTx, groupId, userId, seq, true)
		if err == db.MongoErrorNoNotification {
			retBuf, err = errorHandlerHook(IllegalRequest, "无相关通知")
			return err
		} else if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		if err = db.AgreeJoinGroup(mysqlTx, noti.SenderId, groupId); err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		sendNoti := entities.NewNotification(groupId, noti.SenderId, entities.Approve, entities.Group, true, true)
		sendNoti, err = db.AddAndReturnNotification(mongoTx, sendNoti)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		if _, err = controllers.CheckIsUserInGroup(noti.SenderId, groupId, true); err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		SendNotification(sendNoti)
		SendNotification(noti)
		return nil
	})

	return
}

func refuseJoinGroupRequest(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId  = ctx.Param[userIdFromTokenKey].(int64)
		groupId = ctx.Param[notificationReceiverKey].(int64)
		seq     = ctx.Param[notificationSeqKey].(uint64)
		isAuth  = ctx.Param[groupMemberIsAuthKey].(bool)
	)

	if !isAuth {
		retBuf, err = errorHandlerHook(GroupOpNoAuth, "无权限")
		return
	}

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		var noti *entities.Notification
		noti, err = db.HandleNotification(context.Background(), groupId, userId, seq, false)
		if err == db.MongoErrorNoNotification {
			retBuf, err = errorHandlerHook(IllegalRequest, "无相关通知")
			return err
		} else if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		sendNoti := entities.NewNotification(groupId, noti.SenderId, entities.Refuse, entities.Group, true, false)
		sendNoti, err = db.AddAndReturnNotification(mongoTx, sendNoti)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		SendNotification(sendNoti)
		SendNotification(noti)
		return nil
	})

	return
}

func addAdministrator(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	return addOrDeleteAdministrator(ctx, true)
}

func deleteAdministrator(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	return addOrDeleteAdministrator(ctx, false)
}

func quitOrDeleteMemberFromGroup(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId   = ctx.Param[userIdFromTokenKey].(int64)
		groupId  = ctx.Param[groupIdKey].(int64)
		friendId = ctx.Param[friendIdKey].(int64)
		isAuth   = ctx.Param[groupMemberIsAuthKey].(bool)
		isSame   = ctx.Param[isSameUserKey].(bool)
	)

	if !isAuth && !isSame {
		retBuf, err = errorHandlerHook(IllegalRequest, "非法操作")
		return
	}

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		flag := false
		flag, err = controllers.CheckIsUserInGroup(friendId, groupId, true)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		} else if !flag {
			retBuf, err = errorHandlerHook(UserNotFound, "用户不在群组中")
			return err
		}

		err = db.DeleteFromGroup(mysqlTx, friendId, groupId)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		noti := entities.NewNotification(friendId, groupId, entities.Delete, entities.Group, true, true)
		noti.HandleUserId = userId
		noti, err = db.AddAndReturnNotification(mongoTx, noti)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		_, err = controllers.CheckIsUserInGroup(friendId, groupId, true)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}
		SendNotification(noti)
		return nil
	})

	return
}

func addOrDeleteAdministrator(ctx *controllers.ProcessContext, isAdd bool) (retBuf []byte, err error) {
	var (
		userId   = ctx.Param[userIdFromTokenKey].(int64)
		groupId  = ctx.Param[groupIdKey].(int64)
		friendId = ctx.Param[friendIdKey].(int64)
		isOwner  = ctx.Param[groupMemberIsOwnerKey].(bool)
	)

	if !isOwner {
		retBuf, err = errorHandlerHook(GroupOpNoAuth, "无权限")
		return
	}

	flag, err := controllers.CheckIsUserInGroup(friendId, groupId, true)
	if err != nil {
		retBuf, err = errorHandlerHook(InternalError, err.Error())
		return
	}

	if !flag {
		retBuf, err = errorHandlerHook(UserNotFound, "用户不在群组中")
		return
	}

	db.StartDbTransaction(func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error {
		if isAdd {
			if err = db.AddAdministrator(mysqlTx, friendId, groupId); err == db.MysqlErrorGroupNotExist {
				retBuf, err = errorHandlerHook(IllegalRequest, "用户不存在")
				return err
			} else if err != nil {
				retBuf, err = errorHandlerHook(InternalError, err.Error())
				return err
			}
		} else {
			if err = db.DeleteAdministrator(mysqlTx, friendId, groupId); err == db.MysqlErrorGroupNotExist {
				retBuf, err = errorHandlerHook(IllegalRequest, "用户不存在")
				return err
			} else if err != nil {
				retBuf, err = errorHandlerHook(InternalError, err.Error())
				return err
			}
		}

		opType := entities.Delete
		if isAdd {
			opType = entities.Add
		}

		noti := entities.NewNotification(userId, groupId, opType, entities.Administrator, true, true)
		noti.HandleUserId = userId
		noti, err = db.AddAndReturnNotification(mongoTx, noti)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}

		_, err = controllers.CheckIsUserInGroup(friendId, groupId, true)
		if err != nil {
			retBuf, err = errorHandlerHook(InternalError, err.Error())
			return err
		}
		SendNotification(noti)
		return nil
	})
	return
}

func returnSuccessBody(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	err = (&SuccessBody{ResponseHeader{Success}}).UnmarshalJSON(retBuf)
	return
}

func returnUserInfoBody(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId = ctx.Param[userIdKey].(int64)
		isSelf = ctx.Param[isSameUserKey].(bool)
	)

	info, err := db.SearchUserInfo(nil, userId, isSelf)
	if err == db.MysqlErrorUserNotExist {
		retBuf, err = errorHandlerHook(UserNotFound, "目标用户不存在")
		return
	} else if err != nil {
		return
	}

	err = (&UserInfoBody{
		ResponseHeader: ResponseHeader{Success},
		UserInfo:       *info,
	}).UnmarshalJSON(retBuf)
	return
}

func returnFriendshipBody(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		userId   = ctx.Param[userIdFromTokenKey].(int64)
		friendId = ctx.Param[friendIdKey].(int64)
		chatId   = ctx.Param[chatIdKey].(int64)
	)

	err = (&FriendshipBody{
		ResponseHeader: ResponseHeader{Success},
		Friendship: entities.Friendship{
			GormModel: gorm.Model{},
			SelfId:    userId,
			FriendId:  friendId,
			IsDeleted: false,
			ChatId:    chatId,
		},
	}).UnmarshalJSON(retBuf)
	return
}

func returnGroupInfoBody(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		groupId           = ctx.Param[groupIdKey].(int64)
		groupName         = ctx.Param[groupNameKey].(string)
		groupIntroduction = ctx.Param[groupIntroductionKey].(string)
		groupAvatar       = ctx.Param[groupAvatarKey].(string)
		userId            = ctx.Param[userIdFromTokenKey].(int64)
	)

	err = (&GroupInfoBody{
		ResponseHeader: ResponseHeader{200},
		GroupInfo: entities.GroupInfo{
			Id:           groupId,
			Owner:        userId,
			Name:         groupName,
			Introduction: groupIntroduction,
			Avatar:       groupAvatar,
			IsDeleted:    false,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			Members: []entities.GroupMember{{
				GormModel:       gorm.Model{},
				GroupId:         groupId,
				MemberId:        userId,
				IsAdministrator: true,
				IsDeleted:       false,
			}},
		},
	}).UnmarshalJSON(retBuf)
	return
}

func returnRegisterOrLoginBody(ctx *controllers.ProcessContext) (retBuf []byte, err error) {
	var (
		token  = ctx.Param[tokenKey].(string)
		userId = ctx.Param[userIdKey].(int64)
	)

	err = (&RegisterOrLoginBody{
		ResponseHeader: ResponseHeader{Success},
		Token:          token,
		UserId:         userId,
	}).UnmarshalJSON(retBuf)
	return
}

func getParamFromURL(ctx *controllers.ProcessContext, queryName, errorInfo string, errorStatus int32) (param string, retBuf []byte, err error) {
	ginCtx := ctx.Ctx.(*gin.Context)
	param = ginCtx.Query(queryName)
	if param == "" {
		retBuf, err = errorHandlerHook(errorStatus, errorInfo)
		return
	}
	return
}

func getInt64ParamFromURL(ctx *controllers.ProcessContext, queryName, errorInfo string, errorStatus int32) (param int64, retBuf []byte, err error) {
	tmp := ""
	tmp, retBuf, err = getParamFromURL(ctx, queryName, errorInfo, errorStatus)
	if err != nil {
		return
	}

	param, err = strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		retBuf, err = errorHandlerHook(IllegalRequestFromMismatched, "整数转换无效")
	}
	return
}

func errorHandlerHook(statusCode int32, reason string) (retBuf []byte, err error) {
	log.Error(fmt.Sprintf("处理客户端请求错误。状态码: %d, 错误原因: %s", statusCode, reason))
	if statusCode == InternalError {
		reason = "服务器内部错误"
	}

	err = (&FailBody{
		ResponseHeader: ResponseHeader{statusCode},
		Reason:         reason,
	}).UnmarshalJSON(retBuf)
	if err != nil {
		log.Error(fmt.Sprintf("序列化错误消息发绳错误: %s", err.Error()))
	}
	return
}

// TODO 完善用户注册信息鉴别
func validate(account, email, password string) bool {
	return true
}
