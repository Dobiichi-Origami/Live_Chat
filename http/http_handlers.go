package http

import (
	"github.com/gin-gonic/gin"
	"liveChat/controllers"
)

var (
	loginProcessChain = controllers.NewProcessChain().
				Add(getAccountFromUrl).
				Add(getPasswordFromUrl).
				Add(getUserIdByAccountAndPassword).
				Add(getTokenByUserId).
				Add(returnRegisterOrLoginBody)

	registerProcessChain = controllers.NewProcessChain().
				Add(getAccountFromUrl).
				Add(getPasswordFromUrl).
				Add(getEmailFromUrl).
				Add(validateRegisterInfo).
				Add(registerUserInfoAndGetUserId).
				Add(getTokenByUserId).
				Add(returnRegisterOrLoginBody)

	getUserInfoProcessChain = controllers.NewProcessChain().
				Add(getTokenFromHeader).
				Add(getUserIdFromUrl).
				Add(validateToken).
				Add(tellIsSameUserCompareTokenAndUserId).
				Add(returnUserInfoBody)

	updateUserNameProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getUsernameFromUrl).
					Add(validateToken).
					Add(updateUsername).
					Add(returnSuccessBody)

	updateUserIntroductionProcessChain = controllers.NewProcessChain().
						Add(getTokenFromHeader).
						Add(getUserIntroductionFromUrl).
						Add(validateToken).
						Add(updateUserIntroduction).
						Add(returnSuccessBody)

	updateUserAvatarProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getUserAvatarFromUrl).
					Add(validateToken).
					Add(updateUserAvatar).
					Add(returnSuccessBody)

	addFriendProcessChain = controllers.NewProcessChain().
				Add(getTokenFromHeader).
				Add(getFriendIdFromUrl).
				Add(validateToken).
				Add(tellIsSameUserCompareTokenAndFriendId).
				Add(rejectRequestFromOneSelf).
				Add(sendAddFriendNotificationToOther).
				Add(returnSuccessBody)

	approveFriendApplicationProcessChain = controllers.NewProcessChain().
						Add(getTokenFromHeader).
						Add(getNotificationSeqFromUrl).
						Add(validateToken).
						Add(approveFriendRequest).
						Add(returnFriendshipBody)

	refuseFriendApplicationProcessChain = controllers.NewProcessChain().
						Add(getTokenFromHeader).
						Add(getNotificationSeqFromUrl).
						Add(validateToken).
						Add(refuseFriendRequest).
						Add(returnSuccessBody)

	deleteFriendshipProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getFriendIdFromUrl).
					Add(validateToken).
					Add(tellIsSameUserCompareTokenAndFriendId).
					Add(rejectRequestFromOneSelf).
					Add(sendDeleteFriendNotificationToOther).
					Add(returnSuccessBody)

	getGroupInfoProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getGroupIdFromUrl).
					Add(validateToken).
					Add(getGroupInfoBody)

	createGroupProcessChain = controllers.NewProcessChain().
				Add(getTokenFromHeader).
				Add(getCreateGroupPostInBody).
				Add(validateToken).
				Add(createGroup).
				Add(returnGroupInfoBody)

	deleteGroupProcessChain = controllers.NewProcessChain().
				Add(getTokenFromHeader).
				Add(getGroupIdFromUrl).
				Add(validateToken).
				Add(checkGroupAuth).
				Add(deleteGroup).
				Add(returnSuccessBody)

	updateGroupNameProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getGroupIdFromUrl).
					Add(getGroupNameFromUrl).
					Add(validateToken).
					Add(checkGroupAuth).
					Add(updateGroupName).
					Add(returnSuccessBody)

	updateGroupIntroductionProcessChain = controllers.NewProcessChain().
						Add(getTokenFromHeader).
						Add(getGroupIdFromUrl).
						Add(getGroupIntroductionFromUrl).
						Add(validateToken).
						Add(checkGroupAuth).
						Add(updateGroupIntroduction).
						Add(returnSuccessBody)

	updateGroupAvatarProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getGroupIdFromUrl).
					Add(getGroupAvatarFromUrl).
					Add(validateToken).
					Add(checkGroupAuth).
					Add(updateGroupAvatar).
					Add(returnSuccessBody)

	joinGroupProcessChain = controllers.NewProcessChain().
				Add(getTokenFromHeader).
				Add(getGroupIdFromUrl).
				Add(validateToken).
				Add(checkGroupAuth).
				Add(sendJoinGroupNotification).
				Add(returnSuccessBody)

	approveJoinGroupApplicationProcessChain = controllers.NewProcessChain().
						Add(getTokenFromHeader).
						Add(getGroupIdFromUrl).
						Add(getNotificationSeqFromUrl).
						Add(validateToken).
						Add(checkGroupAuth).
						Add(getGroupIdAsNotificationReceiver).
						Add(approveJoinGroupRequest).
						Add(returnSuccessBody)

	refuseJoinGroupApplicationProcessChain = controllers.NewProcessChain().
						Add(getTokenFromHeader).
						Add(getGroupIdFromUrl).
						Add(getNotificationSeqFromUrl).
						Add(validateToken).
						Add(checkGroupAuth).
						Add(getGroupIdAsNotificationReceiver).
						Add(refuseJoinGroupRequest).
						Add(returnSuccessBody)

	quitOrDeleteMemberProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getGroupIdFromUrl).
					Add(getFriendIdFromUrl).
					Add(validateToken).
					Add(checkGroupAuth).
					Add(tellIsSameUserCompareTokenAndFriendId).
					Add(quitOrDeleteMemberFromGroup).
					Add(returnSuccessBody)

	addAdministratorProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getGroupIdFromUrl).
					Add(getFriendIdFromUrl).
					Add(validateToken).
					Add(checkGroupAuth).
					Add(tellIsSameUserCompareTokenAndFriendId).
					Add(rejectRequestFromOneSelf).
					Add(addAdministrator).
					Add(returnSuccessBody)

	deleteAdministratorProcessChain = controllers.NewProcessChain().
					Add(getTokenFromHeader).
					Add(getGroupIdFromUrl).
					Add(getFriendIdFromUrl).
					Add(validateToken).
					Add(checkGroupAuth).
					Add(tellIsSameUserCompareTokenAndFriendId).
					Add(rejectRequestFromOneSelf).
					Add(deleteAdministrator).
					Add(returnSuccessBody)
)

func init() {

}

func loginHandler(ctx *gin.Context) {
	loginProcessChain.Process(ctx, postHandler)
}

func registerHandler(ctx *gin.Context) {
	registerProcessChain.Process(ctx, postHandler)
}

func getUserInfoHandler(ctx *gin.Context) {
	getUserInfoProcessChain.Process(ctx, postHandler)
}

func updateUserNameHandler(ctx *gin.Context) {
	updateUserNameProcessChain.Process(ctx, postHandler)
}

func updateUserIntroductionHandler(ctx *gin.Context) {
	updateUserIntroductionProcessChain.Process(ctx, postHandler)
}

func updateUserAvatarHandler(ctx *gin.Context) {
	updateUserAvatarProcessChain.Process(ctx, postHandler)
}

func addFriendHandler(ctx *gin.Context) {
	addFriendProcessChain.Process(ctx, postHandler)
}

func deleteFriendHandler(ctx *gin.Context) {
	deleteFriendshipProcessChain.Process(ctx, postHandler)
}

func approveFriendshipApplicationHandler(ctx *gin.Context) {
	approveFriendApplicationProcessChain.Process(ctx, postHandler)
}

func refuseFriendshipApplicationHandler(ctx *gin.Context) {
	refuseFriendApplicationProcessChain.Process(ctx, postHandler)
}

func getGroupInfoHandler(ctx *gin.Context) {
	getGroupInfoProcessChain.Process(ctx, postHandler)
}

func createGroupHandler(ctx *gin.Context) {
	createGroupProcessChain.Process(ctx, postHandler)
}

func deleteGroupHandler(ctx *gin.Context) {
	deleteGroupProcessChain.Process(ctx, postHandler)
}

func updateGroupNameHandler(ctx *gin.Context) {
	updateGroupNameProcessChain.Process(ctx, postHandler)
}

func updateGroupIntroductionHandler(ctx *gin.Context) {
	updateGroupIntroductionProcessChain.Process(ctx, postHandler)
}

func updateGroupAvatarHandler(ctx *gin.Context) {
	updateGroupAvatarProcessChain.Process(ctx, postHandler)
}

func joinGroupHandler(ctx *gin.Context) {
	joinGroupProcessChain.Process(ctx, postHandler)
}

func approveJoinGroupApplicationHandler(ctx *gin.Context) {
	approveJoinGroupApplicationProcessChain.Process(ctx, postHandler)
}

func refuseJoinGroupApplicationHandler(ctx *gin.Context) {
	refuseJoinGroupApplicationProcessChain.Process(ctx, postHandler)
}

func addAdministratorHandler(ctx *gin.Context) {
	addAdministratorProcessChain.Process(ctx, postHandler)
}

func deleteAdministratorHandler(ctx *gin.Context) {
	deleteAdministratorProcessChain.Process(ctx, postHandler)
}

func quitOrDeleteMemberHandler(ctx *gin.Context) {
	quitOrDeleteMemberProcessChain.Process(ctx, postHandler)
}
