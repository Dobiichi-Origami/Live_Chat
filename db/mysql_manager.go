package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"liveChat/config"
	"liveChat/log"
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

	mysqlCreateDatabase         = "CREATE DATABASE IF NOT EXISTS message_server"
	mysqlCreateLoginTable       = "CREATE TABLE IF NOT EXISTS `login` (\n  `userId` bigint(255) NOT NULL,\n  `account` varchar(16) NOT NULL DEFAULT '',\n  `email` varchar(32) NOT NULL DEFAULT '',\n  `password` varchar(32) NOT NULL DEFAULT '',\n  PRIMARY KEY (`id`),\n  UNIQUE KEY `user_id_index` (`account`, `password`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateUserInfoTable    = "CREATE TABLE IF NOT EXISTS `user_info` (\n  `user_id` bigint(20) NOT NULL,\n  `user_name` varchar(32) NOT NULL DEFAULT '',\n  `user_avatar` varchar(1024) NOT NULL DEFAULT '',\n  `user_instruction` text,\n  PRIMARY KEY (`user_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateGroupInfoTable   = "CREATE TABLE IF NOT EXISTS `group_info` (\n  `group_id` bigint(20) NOT NULL,\n  `group_name` varchar(32) NOT NULL DEFAULT '',\n  `owner_id` bigint(20) NOT NULL,\n  `group_instruction` text,\n  `group_avatar` varchar(1024) NOT NULL DEFAULT '',\n  PRIMARY KEY (`group_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateFriendshipTable  = "CREATE TABLE IF NOT EXISTS `friendship` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `master_id` bigint(20) NOT NULL,\n  `slave_id` bigint(20) NOT NULL,\n  `chat_id` bigint(20) NOT NULL,\n  `is_deleted` tinyint(1) NOT NULL,\n  PRIMARY KEY (`id`),\n  KEY `friend_index` (`master_id`,`slave_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	mysqlCreateGroupMemberTable = "CREATE TABLE IF NOT EXISTS `group_member` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `group_id` bigint(20) NOT NULL,\n  `member_id` bigint(20) NOT NULL,\n  `is_administrator` tinyint(1) NOT NULL,\n  `is_deleted` tinyint(1) NOT NULL,\n  PRIMARY KEY (`id`),\n  KEY `group_info_index` (`group_id`,`member_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"

	mysqlDropDatabase = "DROP DATABASE IF EXISTS message_server"
	dropMysqlTable    = "DROP TABLE IF EXISTS `%s`"

	mysqlLoginQuery    = "SELECT `userId` from `login` where `account`=? and `password`=?"
	mysqlRegisterQuery = "INSERT INTO `login` (`userId`, `account`, `email`, `password`) VALUES(?, ?, ?, ?)"
)

var (
	mysqlCfg *config.MysqlConfig = nil

	mysqlDb *sql.DB = nil

	loginStatement    *sql.Stmt = nil
	registerStatement *sql.Stmt = nil

	isMysqlInitiated bool = false
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

	initStatement()

	isMysqlInitiated = true
	return nil
}

func Login(account, password string) (int64, error) {
	rows, err := loginStatement.Query(account, password)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	if !rows.Next() {
		return -1, nil
	}
	var id int64 = -1
	if err = rows.Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

func Register(account, email, password string) (int64, error) {
	if !validateRegisterInfo(account, email, password) {
		return -1, nil
	}

	userId := tools.GenerateSnowflakeId(false)

	_, err := registerStatement.Exec(userId, account, email, password)
	if err != nil {
		// TODO 待优化日志项
		log.Error(err.Error())
		return -1, err
	}

	return userId, nil
}

// TODO 完善用户注册信息鉴别
func validateRegisterInfo(account, email, password string) bool {
	return true
}

func initStatement() {
	var err error

	loginStatement, err = mysqlDb.Prepare(fmt.Sprintf(mysqlLoginQuery, mysqlCfg.UserInfoTable))
	if err != nil {
		// TODO 待优化日志项
		log.Error(err.Error())
		panic(err)
	}

	registerStatement, err = mysqlDb.Prepare(fmt.Sprintf(mysqlRegisterQuery, mysqlCfg.UserInfoTable))
	if err != nil {
		// TODO 待优化日志项
		log.Error(err.Error())
		panic(err)
	}
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

func dropTable(tableName string) error {
	if _, err := mysqlDb.Exec(fmt.Sprintf(dropMysqlTable, tableName)); err != nil {
		return err
	}

	return nil
}
