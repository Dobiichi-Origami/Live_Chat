package http

import (
	"github.com/gin-gonic/gin"
)

var httpServer *gin.Engine

const (
	loginRoute    = "/login"
	registerRoute = "/register"

	userRouteHead = "/userInfo"

	getUserInfoRoute              = userRouteHead
	updateUsernameRoute           = userRouteHead + "/updateUsername"
	updateUserAvatarRoute         = userRouteHead + "/updateUserAvatar"
	updateUserIntroductionRoute   = userRouteHead + "/updateUserIntroduction"
	addFriendRoute                = userRouteHead + "/addFriend"
	approveFriendApplicationRoute = userRouteHead + "/approveFriendApplication"
	refuseFriendApplicationRoute  = userRouteHead + "/refuseFriendApplication"
	deleteFriendRoute             = userRouteHead + "/deleteFriend"

	groupRouteHead = "/groupInfo"

	getGroupInfoRoute            = groupRouteHead
	createGroupRoute             = groupRouteHead + "/createGroup"
	deleteGroupRoute             = groupRouteHead + "/deleteGroup"
	updateGroupNameRoute         = groupRouteHead + "/updateGroupName"
	updateGroupIntroductionRoute = groupRouteHead + "/updateGroupIntroduction"
	updateGroupAvatarRoute       = groupRouteHead + "/updateGroupAvatar"
	joinGroupRoute               = groupRouteHead + "/joinGroup"
	approveJoinApplicationRoute  = groupRouteHead + "/approveJoinApplication"
	refuseJoinApplicationRoute   = groupRouteHead + "/refuseJoinApplication"
	addAdministratorRoute        = groupRouteHead + "/addAdministrator"
	deleteAdministratorRoute     = groupRouteHead + "/deleteAdministrator"
	quitOrDeleteMemberRoute      = groupRouteHead + "/quitOrDeleteMember"
)

func InitHttpServer(addresses []string) {
	httpServer = gin.Default()
	httpServer.GET(loginRoute, loginHandler)
	httpServer.GET(registerRoute, registerHandler)
	httpServer.GET(getUserInfoRoute, getUserInfoHandler)
	httpServer.GET(updateUsernameRoute, updateUserNameHandler)
	httpServer.GET(updateUserAvatarRoute, updateUserAvatarHandler)
	httpServer.GET(updateUserIntroductionRoute, updateUserIntroductionHandler)
	httpServer.GET(addFriendRoute, addFriendHandler)
	httpServer.GET(approveFriendApplicationRoute, approveFriendshipApplicationHandler)
	httpServer.GET(refuseFriendApplicationRoute, refuseFriendshipApplicationHandler)
	httpServer.GET(deleteFriendRoute, deleteFriendHandler)
	httpServer.GET(getGroupInfoRoute, getGroupInfoHandler)
	httpServer.POST(createGroupRoute, createGroupHandler)
	httpServer.GET(deleteGroupRoute, deleteGroupHandler)
	httpServer.GET(updateGroupNameRoute, updateGroupNameHandler)
	httpServer.GET(updateGroupIntroductionRoute, updateGroupIntroductionHandler)
	httpServer.GET(updateGroupAvatarRoute, updateGroupAvatarHandler)
	httpServer.GET(joinGroupRoute, joinGroupHandler)
	httpServer.GET(approveJoinApplicationRoute, approveJoinGroupApplicationHandler)
	httpServer.GET(refuseJoinApplicationRoute, refuseJoinGroupApplicationHandler)
	httpServer.GET(addAdministratorRoute, addAdministratorHandler)
	httpServer.GET(deleteAdministratorRoute, deleteAdministratorHandler)
	httpServer.GET(quitOrDeleteMemberRoute, quitOrDeleteMemberHandler)

	err := httpServer.Run(addresses...)
	if err != nil {
		panic(err)
	}
}
