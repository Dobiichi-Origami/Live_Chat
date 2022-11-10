package db

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"liveChat/config"
	"liveChat/entities"
	"liveChat/tools"
)

const defaultMysqlConfigPath = "./mysql_config.json"

var MysqlConfigPath = defaultMongoDBConfigPath

const (
	mysqlDatabaseName = "message_server"

	mysqlLoginTableName       = "login"
	mysqlUserInfoTableName    = "user_info"
	mysqlGroupInfoTableName   = "group_info"
	mysqlFriendshipTableName  = "friendship"
	mysqlGroupMemberTableName = "group_member"

	mysqlCreateDatabase         = "CREATE DATABASE IF NOT EXISTS message_server;"
	mysqlCreateLoginTable       = "CREATE TABLE IF NOT EXISTS `login` (\n  `userId` bigint(255) NOT NULL,\n  `account` varchar(16) NOT NULL DEFAULT '',\n  `email` varchar(32) NOT NULL DEFAULT '',\n  `password` varchar(32) NOT NULL DEFAULT '',\n  PRIMARY KEY (`userId`),\n  UNIQUE KEY `user_id_index` (`account`, `password`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateUserInfoTable    = "CREATE TABLE IF NOT EXISTS `user_info` (\n  `user_id` bigint(20) NOT NULL,\n  `user_name` varchar(32) NOT NULL DEFAULT '',\n  `user_avatar` varchar(1024) NOT NULL DEFAULT '',\n  `user_introduction` text  NOT NULL,\n  PRIMARY KEY (`user_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateGroupInfoTable   = "CREATE TABLE IF NOT EXISTS `group_info` (\n  `group_id` bigint(20) NOT NULL,\n  `group_name` varchar(32) NOT NULL DEFAULT '',\n  `owner_id` bigint(20) NOT NULL,\n  `group_introduction` text  NOT NULL,\n  `group_avatar` varchar(1024) NOT NULL DEFAULT '',\n  `is_deleted` tinyint(1) NOT NULL,\n  PRIMARY KEY (`group_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateFriendshipTable  = "CREATE TABLE IF NOT EXISTS `friendship` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `master_id` bigint(20) NOT NULL,\n  `slave_id` bigint(20) NOT NULL,\n  `chat_id` bigint(20) NOT NULL,\n  `is_deleted` tinyint(1) NOT NULL,\n  PRIMARY KEY (`id`),\n  KEY `friend_index` (`master_id`,`slave_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateGroupMemberTable = "CREATE TABLE IF NOT EXISTS `group_member` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `group_id` bigint(20) NOT NULL,\n  `member_id` bigint(20) NOT NULL,\n  `is_administrator` tinyint(1) NOT NULL,\n  `is_deleted` tinyint(1) NOT NULL,\n  PRIMARY KEY (`id`),\n  KEY `group_info_index` (`group_id`,`member_id`,`is_deleted`),\n  KEY `member_index` (`member_id`,`is_deleted`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"

	mysqlDropDatabase         = "DROP DATABASE IF EXISTS message_server;"
	dropMysqlTableLogin       = "DROP TABLE IF EXISTS `login`"
	dropMysqlTableUserInfo    = "DROP TABLE IF EXISTS `user_info`"
	dropMysqlTableGroupInfo   = "DROP TABLE IF EXISTS `group_info`"
	dropMysqlTableFriendship  = "DROP TABLE IF EXISTS `friendship`"
	dropMysqlTableGroupMember = "DROP TABLE IF EXISTS `group_member`"

	mysqlLoginQuery       = "SELECT `userId` from `login` where `account`=? and `password`=?;"
	mysqlIsUserExistQuery = "SELECT `userId` from `login` where `account`=?;"
	mysqlRegisterInsert   = "INSERT INTO `login` (`userId`, `account`, `email`, `password`) VALUES(?, ?, ?, ?);"

	mysqlUserInfoQuery                  = "SELECT * from `user_info` where `user_id`=?;"
	mysqlUserInfoInsert                 = "INSERT INTO `user_info` (`user_id`, `user_name`, `user_avatar`, `user_introduction`) VALUES(?, ?, ?, ?);"
	mysqlUserInfoUserNameUpdate         = "UPDATE `user_info` SET `user_name`=? where `user_id`=?;"
	mysqlUserInfoUserAvatarUpdate       = "UPDATE `user_info` SET `user_avatar`=? where `user_id`=?;"
	mysqlUserInfoUserIntroductionUpdate = "UPDATE `user_info` SET `user_introduction`=? where `user_id`=?;"

	mysqlFriendshipOneWayShipInsert      = "INSERT INTO `friendship` (`master_id`, `slave_id`, `chat_id`, `is_deleted`) VALUES(?, ?, ?, 0);"
	mysqlFriendshipBatchOneWayShipQuery  = "SELECT * from (SELECT `master_id`, `slave_id`, `chat_id`, `is_deleted` from `friendship` where `master_id`=?) s where s.is_deleted=0;"
	mysqlFriendshipSingleOneWayShipQuery = "SELECT * from (SELECT `master_id`, `slave_id`, `chat_id`, `is_deleted` from `friendship` where `master_id`=? and `slave_id`=?) s where s.is_deleted=0;"
	mysqlFriendshipOneWayShipDelete      = "UPDATE `friendship` SET `is_deleted`=1 where `master_id`=? and `slave_id`=?;"

	mysqlGroupInfoInsert             = "INSERT INTO `group_info` (`group_id`, `group_name`, `owner_id`, `group_introduction`, `group_avatar`, `is_deleted`) VALUES(?, ?, ?, ?, ?, 0);"
	mysqlGroupInfoQuery              = "SELECT * from `group_info` where `group_id`=? and is_deleted=0;"
	mysqlGroupInfoNameUpdate         = "UPDATE `group_info` SET `group_name`=? where `group_id`=?;"
	mysqlGroupInfoIntroductionUpdate = "UPDATE `group_info` SET `group_introduction`=? where `group_id`=?;"
	mysqlGroupInfoAvatarUpdate       = "UPDATE `group_info` SET `group_avatar`=? where `group_id`=?;"
	mysqlGroupInfoDelete             = "UPDATE `group_info` SET `is_deleted`=1 where `group_id`=?;"

	mysqlGroupMemberInsert                 = "INSERT INTO `group_member` (`group_id`, `member_id`, `is_administrator`, `is_deleted`) VALUES(?, ?, ?, 0);"
	mysqlGroupMemberSelect                 = "SELECT * from (SELECT * from `group_member` where `group_id`=?) s where s.is_deleted=0"
	mysqlGroupMemberReverseSelect          = "SELECT * from `group_member` where `member_id`=? and is_deleted=0;"
	mysqlGroupMemberAdministratorUpdate    = "UPDATE `group_member` SET `is_administrator`=1 where `group_id`=? and `member_id`=?;"
	mysqlGroupMemberNonAdministratorUpdate = "UPDATE `group_member` SET `is_administrator`=0 where `group_id`=? and `member_id`=?;"
	mysqlGroupMemberDelete                 = "UPDATE `group_member` SET `is_deleted`=1 where `group_id`=? and `member_id`=?;"
	mysqlGroupMemberBatchDelete            = "UPDATE `group_member` SET `is_deleted`=1 where `group_id`=?;"
)

var (
	MysqlErrorUserNotExist  = errors.New("用户不存在")
	MysqlErrorGroupNotExist = errors.New("群组不存在")
	MysqlErrorNoLine        = errors.New("无行被插入或修改")
)

var (
	mysqlCfg *config.MysqlConfig = nil
	mysqlDb  *sql.DB             = nil

	isMysqlInitiated bool = false
)

var (
	loginStatement          *sql.Stmt
	isUserExistStatement    *sql.Stmt
	insertRegisterStatement *sql.Stmt
	insertUserInfoStatement *sql.Stmt

	selectUserInfoStatement         *sql.Stmt
	updateUserNameStatement         *sql.Stmt
	updateUserAvatarStatement       *sql.Stmt
	updateUserIntroductionStatement *sql.Stmt

	insertOneWayFriendshipStatement       *sql.Stmt
	selectSingleOneWayFriendShipStatement *sql.Stmt
	selectBatchOneWayFriendShipStatement  *sql.Stmt
	deleteOneWayFriendshipStatement       *sql.Stmt

	insertGroupInfoStatement             *sql.Stmt
	selectGroupInfoStatement             *sql.Stmt
	updateGroupInfoNameStatement         *sql.Stmt
	updateGroupInfoIntroductionStatement *sql.Stmt
	updateGroupInfoAvatarStatement       *sql.Stmt
	deleteGroupInfoStatement             *sql.Stmt

	insertGroupMemberStatement                 *sql.Stmt
	selectGroupMemberListStatement             *sql.Stmt
	selectGroupMemberReverseStatement          *sql.Stmt
	updateGroupMemberAdministratorStatement    *sql.Stmt
	updateGroupMemberNonAdministratorStatement *sql.Stmt
	deleteGroupMemberStatement                 *sql.Stmt
	deleteGroupMemberBatchStatement            *sql.Stmt
)

func InitMysqlConnection(configPath string) error {
	if isMysqlInitiated {
		return nil
	}

	path := tools.GetPath(MysqlConfigPath, configPath)
	mysqlCfg = config.NewMysqlConfig(path)
	url := mysqlCfg.Format()

	tmpDb, err := sql.Open("mysql", url)
	if err != nil {
		return err
	}
	mysqlDb = tmpDb

	err = mysqlDb.Ping()
	if err != nil {
		return err
	}

	if err = createTableWithIndex(mysqlCfg.UserInfoTable); err != nil {
		return err
	}

	if err = initStatement(); err != nil {
		return err
	}

	isMysqlInitiated = true
	return nil
}

func Login(account, password string) (int64, error) {
	rows, err := loginStatement.Query(account, password)
	if err != nil {
		return -1, err
	} else if !rows.Next() {
		return -1, MysqlErrorUserNotExist
	}

	var id = int64(0)
	if err = rows.Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

func Register(account, email, password string) (id int64, err error) {
	if !validateRegisterInfo(account, email, password) {
		// TODO 加入错误
		return -1, nil
	}

	tx, err := mysqlDb.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	rows, err := tx.Stmt(isUserExistStatement).Query(account)
	if err != nil {
		return -1, err
	} else if rows.Next() {
		return -1, MysqlErrorUserNotExist
	}

	userId := tools.GenerateSnowflakeId(false)
	if _, err = tx.Stmt(insertRegisterStatement).Exec(userId, account, email, password); err != nil {
		return -1, err
	}

	if _, err = tx.Stmt(insertUserInfoStatement).Exec(userId, entities.DefaultUserName, entities.DefaultUserAvatar, entities.DefaultUserIntroduction); err != nil {
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}

	return userId, nil
}

func SearchUserInfo(id int64) (*entities.UserInfo, error) {
	rows, err := selectUserInfoStatement.Query(id)
	if err != nil {
		return nil, err
	}

	return entities.ScanUserInfoFromSqlResult(rows)
}

func UpdateUserName(userId int64, userName string) error {
	return resultAndErrorHandle(updateUserNameStatement.Exec(userName, userId))
}

func UpdateUserAvatar(userId int64, userAvatar string) error {
	return resultAndErrorHandle(updateUserAvatarStatement.Exec(userAvatar, userId))
}

func UpdateUserIntroduction(userId int64, userIntroduction string) error {
	return resultAndErrorHandle(updateUserIntroductionStatement.Exec(userIntroduction, userId))
}

func AgreeFriendShip(userId1, userId2 int64) (chatId int64, err error) {
	tx, err := mysqlDb.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	chatId = tools.GenerateSnowflakeId(false)
	if _, err = tx.Stmt(insertOneWayFriendshipStatement).Exec(userId1, userId2, chatId); err != nil {
		return -1, err
	}

	if _, err = tx.Stmt(insertOneWayFriendshipStatement).Exec(userId2, userId1, chatId); err != nil {
		return -1, err
	}

	if err = addChat(context.Background(), chatId); err != nil {
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}

	return chatId, nil
}

func DeleteFriendShip(userId1, userId2 int64) error {
	tx, err := mysqlDb.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.Stmt(deleteOneWayFriendshipStatement).Exec(userId1, userId2); err != nil {
		return err
	}

	if _, err = tx.Stmt(deleteOneWayFriendshipStatement).Exec(userId2, userId1); err != nil {
		return err
	}

	return tx.Commit()
}

func SelectFriendShip(userId int64) ([]entities.Friendship, error) {
	rows, err := selectBatchOneWayFriendShipStatement.Query(userId)
	if err != nil {
		return nil, err
	}

	return entities.ScanFriendshipsFromSqlResult(rows)
}

func AddGroupInfo(owner int64, name, introduction, avatar string) (id int64, err error) {
	tx, err := mysqlDb.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	groupId := tools.GenerateSnowflakeId(true)
	_, err = tx.Stmt(insertGroupInfoStatement).Exec(groupId, name, owner, introduction, avatar)
	if err != nil {
		return -1, err
	}

	_, err = tx.Stmt(insertGroupMemberStatement).Exec(groupId, owner, 1)
	if err != nil {
		return -1, err
	}

	if err = addChat(context.Background(), groupId); err != nil {
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}

	return groupId, nil
}

func SearchGroupInfo(id int64) (*entities.GroupInfo, error) {
	rows, err := selectGroupInfoStatement.Query(id)
	if err != nil {
		return nil, err
	} else if !rows.Next() {
		return nil, MysqlErrorGroupNotExist
	}

	return entities.ScanGroupInfoFromSqlResult(rows)
}

func DeleteGroupInfo(groupId int64) (err error) {
	tx, err := mysqlDb.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.Stmt(deleteGroupInfoStatement).Exec(groupId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return MysqlErrorGroupNotExist
	}

	_, err = tx.Stmt(deleteGroupMemberBatchStatement).Exec(groupId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func AgreeJoinGroup(userId, groupId int64) error {
	return resultAndErrorHandle(insertGroupMemberStatement.Exec(groupId, userId, 0))
}

func DeleteFromGroup(userId, groupId int64) error {
	return resultAndErrorHandle(deleteGroupMemberStatement.Exec(groupId, userId))
}

func AddAdministrator(userId, groupId int64) error {
	return resultAndErrorHandle(updateGroupMemberAdministratorStatement.Exec(groupId, userId))
}

func DeleteAdministrator(userId, groupId int64) error {
	return resultAndErrorHandle(updateGroupMemberNonAdministratorStatement.Exec(groupId, userId))
}

func UpdateGroupName(groupId int64, name string) error {
	return resultAndErrorHandle(updateGroupInfoNameStatement.Exec(name, groupId))
}

func UpdateGroupIntroduction(groupId int64, introduction string) error {
	return resultAndErrorHandle(updateGroupInfoIntroductionStatement.Exec(introduction, groupId))
}

func UpdateGroupAvatar(groupId int64, avatar string) error {
	return resultAndErrorHandle(updateGroupInfoAvatarStatement.Exec(avatar, groupId))
}

func SelectGroupMemberList(groupId int64) ([]entities.GroupMember, error) {
	rows, err := selectGroupMemberListStatement.Query(groupId)
	if err != nil {
		return nil, err
	}

	return entities.ScanGroupMemberFromSqlResult(rows)
}

func SelectGroupInfoForUser(userId int64) ([]entities.GroupMember, error) {
	rows, err := selectGroupMemberReverseStatement.Query(userId)
	if err != nil {
		return nil, err
	}

	return entities.ScanGroupMemberFromSqlResult(rows)
}

// TODO 完善用户注册信息鉴别
func validateRegisterInfo(account, email, password string) bool {
	return true
}

func resultAndErrorHandle(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	cout, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if cout == 0 {
		return MysqlErrorNoLine
	}

	return nil
}

func initStatement() (err error) {
	loginStatement, err = mysqlDb.Prepare(mysqlLoginQuery)
	if err != nil {
		return err
	}
	isUserExistStatement, err = mysqlDb.Prepare(mysqlIsUserExistQuery)
	if err != nil {
		return err
	}
	insertRegisterStatement, err = mysqlDb.Prepare(mysqlRegisterInsert)
	if err != nil {
		return err
	}
	insertUserInfoStatement, err = mysqlDb.Prepare(mysqlUserInfoInsert)
	if err != nil {
		return err
	}

	selectUserInfoStatement, err = mysqlDb.Prepare(mysqlUserInfoQuery)
	if err != nil {
		return err
	}
	updateUserNameStatement, err = mysqlDb.Prepare(mysqlUserInfoUserNameUpdate)
	if err != nil {
		return err
	}
	updateUserAvatarStatement, err = mysqlDb.Prepare(mysqlUserInfoUserAvatarUpdate)
	if err != nil {
		return err
	}
	updateUserIntroductionStatement, err = mysqlDb.Prepare(mysqlUserInfoUserIntroductionUpdate)
	if err != nil {
		return err
	}

	insertOneWayFriendshipStatement, err = mysqlDb.Prepare(mysqlFriendshipOneWayShipInsert)
	if err != nil {
		return err
	}
	selectSingleOneWayFriendShipStatement, err = mysqlDb.Prepare(mysqlFriendshipSingleOneWayShipQuery)
	if err != nil {
		return err
	}
	selectBatchOneWayFriendShipStatement, err = mysqlDb.Prepare(mysqlFriendshipBatchOneWayShipQuery)
	if err != nil {
		return err
	}
	deleteOneWayFriendshipStatement, err = mysqlDb.Prepare(mysqlFriendshipOneWayShipDelete)
	if err != nil {
		return err
	}

	insertGroupInfoStatement, err = mysqlDb.Prepare(mysqlGroupInfoInsert)
	if err != nil {
		return err
	}
	selectGroupInfoStatement, err = mysqlDb.Prepare(mysqlGroupInfoQuery)
	if err != nil {
		return err
	}
	updateGroupInfoNameStatement, err = mysqlDb.Prepare(mysqlGroupInfoNameUpdate)
	if err != nil {
		return err
	}
	updateGroupInfoIntroductionStatement, err = mysqlDb.Prepare(mysqlGroupInfoIntroductionUpdate)
	if err != nil {
		return err
	}
	updateGroupInfoAvatarStatement, err = mysqlDb.Prepare(mysqlGroupInfoAvatarUpdate)
	if err != nil {
		return err
	}
	deleteGroupInfoStatement, err = mysqlDb.Prepare(mysqlGroupInfoDelete)
	if err != nil {
		return err
	}

	insertGroupMemberStatement, err = mysqlDb.Prepare(mysqlGroupMemberInsert)
	if err != nil {
		return err
	}
	selectGroupMemberListStatement, err = mysqlDb.Prepare(mysqlGroupMemberSelect)
	if err != nil {
		return err
	}
	selectGroupMemberReverseStatement, err = mysqlDb.Prepare(mysqlGroupMemberReverseSelect)
	if err != nil {
		return err
	}
	updateGroupMemberAdministratorStatement, err = mysqlDb.Prepare(mysqlGroupMemberAdministratorUpdate)
	if err != nil {
		return err
	}
	updateGroupMemberNonAdministratorStatement, err = mysqlDb.Prepare(mysqlGroupMemberNonAdministratorUpdate)
	if err != nil {
		return err
	}
	deleteGroupMemberStatement, err = mysqlDb.Prepare(mysqlGroupMemberDelete)
	if err != nil {
		return err
	}
	deleteGroupMemberBatchStatement, err = mysqlDb.Prepare(mysqlGroupMemberBatchDelete)
	if err != nil {
		return err
	}

	return err
}

func createTableWithIndex(tableName string) error {
	if _, err := mysqlDb.Exec(mysqlCreateLoginTable); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(mysqlCreateUserInfoTable); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(mysqlCreateGroupInfoTable); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(mysqlCreateFriendshipTable); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(mysqlCreateGroupMemberTable); err != nil {
		return err
	}
	return nil
}

func dropTable() error {
	if _, err := mysqlDb.Exec(dropMysqlTableLogin); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(dropMysqlTableUserInfo); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(dropMysqlTableGroupInfo); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(dropMysqlTableFriendship); err != nil {
		return err
	}

	if _, err := mysqlDb.Exec(dropMysqlTableGroupMember); err != nil {
		return err
	}

	return nil
}
