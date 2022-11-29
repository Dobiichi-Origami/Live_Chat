package db

import (
	"context"
	"database/sql"
	"errors"
	gormSql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"liveChat/config"
	"liveChat/entities"
	"liveChat/tools"
	"time"
)

const defaultMysqlConfigPath = "./mysql_config.json"

var MysqlConfigPath = defaultMongoDBConfigPath

var (
	MysqlErrorUserNotExist  = errors.New("用户不存在")
	MysqlErrorGroupNotExist = errors.New("群组不存在")
	MysqlErrorNoLine        = errors.New("无行被插入或修改")
	MysqlValidateFailed     = errors.New("用户信息校验失败")
)

var (
	mysqlCfg *config.MysqlConfig = nil
	mysqlDb  *gorm.DB            = nil

	isMysqlInitiated bool = false
)

type loginTableEntry struct {
	Id       int64  `gorm:"primaryKey"`
	Account  string `gorm:"uniqueIndex;index:login_index"`
	Password string `gorm:"index:login_index"`
	Email    string
}

func InitMysqlConnection(configPath string) error {
	if isMysqlInitiated {
		return nil
	}

	var err = error(nil)

	path := tools.GetPath(MysqlConfigPath, configPath)
	mysqlCfg = config.NewMysqlConfig(path)
	url := mysqlCfg.Format()

	if mysqlDb, err = gorm.Open(gormSql.Open(url), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}); err != nil {
		return err
	}

	if err = mysqlDb.AutoMigrate(&loginTableEntry{}, &entities.UserInfo{}, &entities.GroupInfo{}, &entities.Friendship{}, &entities.GroupMember{}); err != nil {
		return err
	}

	isMysqlInitiated = true
	return nil
}

func Login(executor *gorm.DB, account, password string) (id int64, err error) {
	executor = returnMysqlDbObj(executor)
	result := executor.Model(&loginTableEntry{}).Select("id").Where("account = ? AND password = ?", account, password).Find(&id)
	if result.Error != nil {
		return -1, result.Error
	} else if result.RowsAffected == 0 {
		return -1, MysqlErrorUserNotExist
	}

	return id, nil
}

func Register(executor *gorm.DB, account, email, password string) (id int64, err error) {
	executor = returnMysqlDbObj(executor)
	err = executor.Transaction(func(tx *gorm.DB) (err error) {

		id = tools.GenerateSnowflakeId(false)
		entry := loginTableEntry{
			Id:       id,
			Account:  account,
			Password: password,
			Email:    email,
		}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}

		if err := tx.Create(entities.NewUserInfoWithDefaultValue(id)).Error; err != nil {
			return err
		}

		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})

	if err != nil {
		return -1, err
	}
	return
}

func SearchUserInfo(executor *gorm.DB, id int64, isSelf bool) (*entities.UserInfo, error) {
	executor = returnMysqlDbObj(executor)
	var (
		info   = entities.NewEmptyUserInfo()
		result *gorm.DB
	)

	if isSelf {
		result = executor.Preload("Groups", "member_id = ? AND is_deleted = 0", id).
			Preload("Friendships", "self_id = ? AND is_deleted = 0", id).
			Where("id = ?", id).
			Find(info)
	} else {
		result = executor.
			Where("id = ?", id).
			Find(info)
	}

	if result.Error != nil {
		return nil, result.Error
	} else if result.RowsAffected != 1 {
		return nil, MysqlErrorUserNotExist
	}

	return info, nil
}

func UpdateUserName(executor *gorm.DB, userId int64, userName string) error {
	executor = returnMysqlDbObj(executor)
	return updateUserInfo(executor, userId, "username", userName)
}

func UpdateUserAvatar(executor *gorm.DB, userId int64, userAvatar string) error {
	executor = returnMysqlDbObj(executor)
	return updateUserInfo(executor, userId, "user_avatar", userAvatar)
}

func UpdateUserIntroduction(executor *gorm.DB, userId int64, userIntroduction string) error {
	executor = returnMysqlDbObj(executor)
	return updateUserInfo(executor, userId, "user_introduction", userIntroduction)
}

func AgreeFriendShip(executor *gorm.DB, userId1, userId2 int64) (chatId int64, err error) {
	executor = returnMysqlDbObj(executor)
	err = executor.Transaction(func(tx *gorm.DB) error {
		if err := isUserInfoExist(tx, userId1); err != nil {
			return err
		}

		if err := isUserInfoExist(tx, userId2); err != nil {
			return err
		}

		chatId = tools.GenerateSnowflakeId(false)
		clauseCond := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "self_id"}, {Name: "friend_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"is_deleted": false}),
		})

		if err := clauseCond.Create(entities.NewFriendship(userId1, userId2, chatId)).Error; err != nil {
			return err
		}
		if err := clauseCond.Create(entities.NewFriendship(userId2, userId1, chatId)).Error; err != nil {
			return err
		}

		info := &entities.Friendship{}
		if err := tx.Where("self_id = ? AND friend_id = ?", userId1, userId2).Find(info).Error; err != nil {
			return err
		}

		chatId = info.ChatId
		return addChat(context.Background(), chatId)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})

	if err != nil {
		return -1, err
	}
	return
}

func DeleteFriendShip(executor *gorm.DB, userId1, userId2 int64) error {
	executor = returnMysqlDbObj(executor)
	return executor.Transaction(func(tx *gorm.DB) error {
		if err := setDeleteFlagForFriendship(tx, userId1, userId2); err != nil {
			return err
		} else if err = setDeleteFlagForFriendship(tx, userId2, userId1); err != nil {
			return err
		}

		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
}

func SelectFriendShip(executor *gorm.DB, userId int64) ([]entities.Friendship, error) {
	executor = returnMysqlDbObj(executor)
	friendships := make([]entities.Friendship, 0)
	if result := executor.Where("self_id = ? AND is_deleted = 0", userId).Find(&friendships); result.Error != nil {
		return nil, result.Error
	}
	return friendships, nil
}

func TellIsFriendBetween(executor *gorm.DB, userId1, userId2 int64) (bool, int64, error) {
	executor = returnMysqlDbObj(executor)
	friendship := &entities.Friendship{}
	ret := executor.Where("self_id = ? AND friend_id = ?", userId1, userId2).Find(friendship)
	if ret.Error != nil {
		return false, 0, ret.Error
	}

	if ret.RowsAffected == 1 {
		return friendship.IsDeleted, friendship.GormModel.UpdatedAt.UnixMilli(), nil
	}

	return false, time.Now().UnixMilli(), nil
}

func AddGroupInfo(executor *gorm.DB, owner int64, name, introduction, avatar string) (id int64, err error) {
	executor = returnMysqlDbObj(executor)
	err = executor.Transaction(func(tx *gorm.DB) error {
		id = tools.GenerateSnowflakeId(true)
		if err := tx.Create(entities.NewGroupInfo(id, owner, name, introduction, avatar)).Error; err != nil {
			return err
		}

		if err := tx.Create(entities.NewGroupMember(id, owner, true)).Error; err != nil {
			return err
		}

		return addChat(context.Background(), id)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})

	if err != nil {
		return -1, err
	}
	return
}

func SearchGroupInfo(executor *gorm.DB, id int64, isInGroup bool) (*entities.GroupInfo, error) {
	executor = returnMysqlDbObj(executor)
	var (
		groupInfo = entities.NewEmptyGroupInfo()
		result    *gorm.DB
	)

	if isInGroup {
		result = executor.Preload("Members", executor.Where(&entities.GroupMember{GroupId: id, IsDeleted: false})).Where("id = ?", id).Find(groupInfo)
	} else {
		result = executor.Where("id = ?", id).Find(groupInfo)
	}

	if result.Error != nil {
		return nil, result.Error
	} else if result.RowsAffected != 1 {
		return nil, MysqlErrorGroupNotExist
	}

	return groupInfo, nil
}

func DeleteGroupInfo(executor *gorm.DB, groupId int64) (err error) {
	executor = returnMysqlDbObj(executor)
	groupInfo, err := SearchGroupInfo(executor, groupId, true)
	if err != nil {
		return err
	}

	groupInfo.IsDeleted = true
	for _, member := range groupInfo.Members {
		member.IsDeleted = true
	}

	return executor.
		Session(&gorm.Session{FullSaveAssociations: true, SkipDefaultTransaction: false}).
		Save(groupInfo).
		Error
}

func AgreeJoinGroup(executor *gorm.DB, userId, groupId int64) error {
	executor = returnMysqlDbObj(executor)
	clauseCond := executor.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "member_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"is_deleted": false}),
	})
	return clauseCond.Create(entities.NewGroupMember(groupId, userId, false)).Error
}

func DeleteFromGroup(executor *gorm.DB, userId, groupId int64) error {
	executor = returnMysqlDbObj(executor)
	return setDeleteFlagForGroupMember(executor, groupId, userId)
}

func AddAdministrator(executor *gorm.DB, userId, groupId int64) error {
	executor = returnMysqlDbObj(executor)
	result := executor.
		Model(&entities.GroupMember{}).
		Where("group_id = ? AND member_id = ? AND is_deleted = 0", groupId, userId).
		Update("is_administrator", true)

	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return MysqlErrorUserNotExist
	}

	return nil
}

func DeleteAdministrator(executor *gorm.DB, userId, groupId int64) error {
	executor = returnMysqlDbObj(executor)
	result := executor.
		Model(&entities.GroupMember{}).
		Where("group_id = ? AND member_id = ? AND is_deleted = 0", groupId, userId).
		Update("is_administrator", false)

	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return MysqlErrorUserNotExist
	}

	return nil
}

func UpdateGroupName(executor *gorm.DB, groupId int64, name string) error {
	executor = returnMysqlDbObj(executor)
	return updateGroupInfo(executor, groupId, "name", name)
}

func UpdateGroupIntroduction(executor *gorm.DB, groupId int64, introduction string) error {
	executor = returnMysqlDbObj(executor)
	return updateGroupInfo(executor, groupId, "introduction", introduction)
}

func UpdateGroupAvatar(executor *gorm.DB, groupId int64, avatar string) error {
	executor = returnMysqlDbObj(executor)
	return updateGroupInfo(executor, groupId, "avatar", avatar)
}

func SelectGroupMemberList(executor *gorm.DB, groupId int64) ([]entities.GroupMember, error) {
	executor = returnMysqlDbObj(executor)
	info, err := SearchGroupInfo(executor, groupId, true)
	if err != nil {
		return nil, err
	}

	return info.Members, nil
}

func SelectGroupInfoForUser(executor *gorm.DB, userId int64) ([]entities.GroupMember, error) {
	executor = returnMysqlDbObj(executor)
	ret := make([]entities.GroupMember, 0)
	result := executor.Model(&entities.GroupMember{}).Where("member_id = ? AND is_deleted = 0", userId).Find(&ret)
	if result.Error != nil {
		return nil, result.Error
	}
	return ret, nil
}

func updateUserInfo(executor *gorm.DB, userId int64, columnName, columnValue string) error {
	if result := executor.Model(&entities.UserInfo{}).Where("id = ?", userId).Update(columnName, columnValue); result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return MysqlErrorUserNotExist
	}
	return nil
}

func isUserInfoExist(tx *gorm.DB, userId int64) error {
	result := tx.Model(&entities.UserInfo{}).Where("id = ?", userId).Find(entities.NewEmptyUserInfo())
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return MysqlErrorUserNotExist
	}

	return nil
}

func setDeleteFlagForFriendship(tx *gorm.DB, userId1, userId2 int64) error {
	result := tx.Model(&entities.Friendship{}).Where("self_id = ? AND friend_id = ?", userId1, userId2).Update("is_deleted", true)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return MysqlErrorUserNotExist
	}

	return nil
}

func setDeleteFlagForGroupMember(tx *gorm.DB, groupId, userId int64) error {
	result := tx.Model(&entities.GroupMember{}).Where("group_id = ? AND member_id = ?", groupId, userId).Update("is_deleted", true)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return MysqlErrorUserNotExist
	}

	return nil
}

func updateGroupInfo(executor *gorm.DB, groupId int64, columnName, columnValue string) error {
	if result := executor.Model(entities.NewEmptyGroupInfo()).Where("id = ?", groupId).Update(columnName, columnValue); result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return MysqlErrorGroupNotExist
	}

	return nil
}

func returnMysqlDbObj(obj *gorm.DB) *gorm.DB {
	if obj != nil {
		return obj
	}
	return mysqlDb
}
