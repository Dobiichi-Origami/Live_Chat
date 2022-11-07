package config

import (
	"encoding/json"
	"io/ioutil"
	"liveChat/log"
	"net/url"
	"strings"
)

type AddressWithPort struct {
	Address  string `json:"address"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type AddressPortInfo []AddressWithPort

func (list AddressPortInfo) writeAddressWithAuth(builder *strings.Builder, authCon, addressCon, addressSep byte, isMysql bool) {
	for index, entry := range list {
		if entry.Username != "" {
			builder.WriteString(url.QueryEscape(entry.Username))
			builder.WriteByte(authCon)
			builder.WriteString(url.QueryEscape(entry.Password))
			builder.WriteByte(addressCon)
		}

		if isMysql {
			builder.WriteString("tcp(")
		}

		builder.WriteString(entry.Address)
		if entry.Port != "" {
			builder.WriteByte(':')
			builder.WriteString(entry.Port)
		}

		if isMysql {
			builder.WriteByte(')')
		}

		if index != len(list)-1 {
			builder.WriteByte(addressSep)
		}
	}
}

// MongoDB Config
const (
	mongodbProtocolHead = "mongodb"
	mongodbDNSHead      = "+srv"
	mongodbProtocolSep  = "://"

	mongodbOptionSep  = '&'
	mongodbOptionCon  = '='
	mongodbAddressSep = ','
	mongodbAddressCon = '@'
	mongodbAuthCon    = ':'
)

type MongoDBConnectionConfig struct {
	AddressList       AddressPortInfo   `json:"address_list"`
	IsDNS             bool              `json:"is_dns,omitempty"`
	AuthDB            string            `json:"auth_db,omitempty"`
	ConnectionOptions map[string]string `json:"connection_options,omitempty"`
}

type MongoDBConfig struct {
	Connection MongoDBConnectionConfig `json:"connection"`

	Database            string `json:"database"`
	MessageCollection   string `json:"message_collection"`
	QueueCollection     string `json:"queue_collection"`
	UserInfoCollection  string `json:"user_info_collection"`
	GroupInfoCollection string `json:"group_info_collection"`
}

func NewMongoDBConfig(path string) *MongoDBConfig {
	config := MongoDBConfig{}
	readConfigFile(path, &config)
	return &config
}

func (cfg *MongoDBConfig) Format() string {
	builder := strings.Builder{}
	builder.Grow(128)
	config := cfg.Connection

	builder.WriteString(mongodbProtocolHead)
	if config.IsDNS {
		builder.WriteString(mongodbDNSHead)
	}
	builder.WriteString(mongodbProtocolSep)

	config.AddressList.writeAddressWithAuth(&builder, mongodbAuthCon, mongodbAddressCon, mongodbAddressSep, false)

	if config.AuthDB != "" || len(config.ConnectionOptions) != 0 {
		builder.WriteByte('/')
	}

	if config.AuthDB != "" {
		builder.WriteString(config.AuthDB)
	}

	if len(config.ConnectionOptions) != 0 {
		builder.WriteByte('?')
		writeOptionsWithSepAndCon(&builder, config.ConnectionOptions, mongodbOptionSep, mongodbOptionCon)
	}
	return builder.String()
}

// Redis Config
const (
	redisProtocolHead = "redis://"
	redisAddressCon   = '@'
	redisAuthCon      = ':'
)

type RedisConfig struct {
	AddressList AddressPortInfo `json:"address_list"`
	Db          string          `json:"db,omitempty"`
}

func NewRedisConfig(path string) *RedisConfig {
	config := RedisConfig{}
	readConfigFile(path, &config)
	return &config
}

func (cfg *RedisConfig) Format() string {
	builder := strings.Builder{}
	builder.Grow(128)

	builder.WriteString(redisProtocolHead)
	cfg.AddressList.writeAddressWithAuth(&builder, redisAuthCon, redisAddressCon, ',', false)

	if cfg.Db != "" {
		builder.WriteByte('/')
		builder.WriteString(cfg.Db)
	}
	return builder.String()
}

// Mysql Config
const (
	mysqlProtocolHead = "mysql://"
	mysqlAddressCon   = '@'
	mysqlAuthCon      = ':'
	mysqlOptionsSep   = '&'
	mysqlOptionsCon   = '='
)

type MysqlConfig struct {
	AddressList       AddressPortInfo   `json:"address_list"`
	ConnectionOptions map[string]string `json:"connection_options,omitempty"`
	Db                string            `json:"db"`
	UserInfoTable     string            `json:"user_info_table"`
}

func NewMysqlConfig(path string) *MysqlConfig {
	config := MysqlConfig{}
	readConfigFile(path, &config)
	return &config
}

func (cfg *MysqlConfig) Format() string {
	builder := strings.Builder{}
	builder.Grow(128)

	cfg.AddressList.writeAddressWithAuth(&builder, mysqlAuthCon, mysqlAddressCon, ',', true)
	builder.WriteString("/")
	builder.WriteString(cfg.Db)

	if len(cfg.ConnectionOptions) != 0 {
		builder.WriteByte('?')
		writeOptionsWithSepAndCon(&builder, cfg.ConnectionOptions, mysqlOptionsSep, mysqlOptionsCon)
	}

	return builder.String()
}

// Tools
func readConfigFile(path string, container interface{}) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		// TODO 待优化日志项
		log.Error(err.Error())
		panic(err)
	}

	err = json.Unmarshal(data, container)
	if err != nil {
		// TODO 待优化日志项
		log.Error(err.Error())
		panic(err)
	}
}

func writeOptionsWithSepAndCon(builder *strings.Builder, configs map[string]string, sep, con byte) {
	counter := 0
	for key, value := range configs {
		counter++
		builder.WriteString(key)
		builder.WriteByte(con)
		builder.WriteString(value)
		if counter != len(configs) {
			builder.WriteByte(sep)
		}
	}
}
