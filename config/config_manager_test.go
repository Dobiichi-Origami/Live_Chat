package config

import (
	"fmt"
	"reflect"
	"testing"
)

const localTestBenchAddress = "192.168.199.235"

var (
	intendedMongoDbConfig = MongoDBConfig{
		Connection: MongoDBConnectionConfig{
			AddressList:       []AddressWithPort{{Address: localTestBenchAddress, Port: "27017"}},
			IsDNS:             false,
			AuthDB:            "",
			ConnectionOptions: nil,
		},
		Database:            "default_database",
		MessageCollection:   "message",
		QueueCollection:     "queue",
		UserInfoCollection:  "user_info",
		GroupInfoCollection: "group_info",
	}

	targetMongodbUrl = fmt.Sprintf("mongodb://%s:27017", localTestBenchAddress)
)

func TestMongoDbConstruction(t *testing.T) {
	if targetMongodbUrl != intendedMongoDbConfig.Format() {
		t.Errorf(formatErrorInfo("generated mongodb connection url mismatched", targetMongodbUrl, intendedMongoDbConfig.Format()))
	}
}

func TestMongoDbConfigRead(t *testing.T) {
	path := "../default_config_files/default_mongodb_config.json"
	config := NewMongoDBConfig(path)

	if !reflect.DeepEqual(*config, intendedMongoDbConfig) {
		t.Errorf(formatErrorInfo("read mongodb config mismatched", intendedMongoDbConfig, *config))
	}
}

var (
	intendedMysqlDbConfig = MysqlConfig{
		AddressList:       []AddressWithPort{{Address: localTestBenchAddress, Port: "3306", Username: "root", Password: "password"}},
		ConnectionOptions: nil,
		Db:                "default_database",
		UserInfoTable:     "user_info",
	}

	targetMysqlUrl = fmt.Sprintf("root:password@tcp(%s:3306)/default_database?parseTime=true", localTestBenchAddress)
)

func TestMysqlConfigRead(t *testing.T) {
	path := "../default_config_files/default_mysql_config.json"
	config := NewMysqlConfig(path)

	if !reflect.DeepEqual(*config, intendedMysqlDbConfig) {
		t.Errorf(formatErrorInfo("read mysql config mismatched", intendedMysqlDbConfig, *config))
	}
}

func TestMysqlConstruction(t *testing.T) {
	if targetMysqlUrl != intendedMysqlDbConfig.Format() {
		t.Errorf(formatErrorInfo("generated mysql connection url mismatched", targetMysqlUrl, intendedMysqlDbConfig.Format()))
	}
}

var (
	intendedRedisConfig = RedisConfig{
		AddressList: []AddressWithPort{{Address: localTestBenchAddress, Port: "6379"}},
		Db:          "0",
	}

	targetRedisUrl = fmt.Sprintf("redis://%s:6379/0", localTestBenchAddress)
)

func TestRedisConstruction(t *testing.T) {
	if targetRedisUrl != intendedRedisConfig.Format() {
		t.Errorf(formatErrorInfo("generated redis connection url mismatched", targetRedisUrl, intendedRedisConfig.Format()))
	}
}

func TestRedisConfigRead(t *testing.T) {
	path := "../default_config_files/default_redis_config.json"
	config := NewRedisConfig(path)

	if !reflect.DeepEqual(*config, intendedRedisConfig) {
		t.Errorf(formatErrorInfo("read redis config mismatched", intendedRedisConfig, *config))
	}
}

func formatErrorInfo(brief string, target, practical interface{}) string {
	return fmt.Sprintf("%s, intended: %+v, actual: %+v", brief, target, practical)
}
