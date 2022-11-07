package db

import (
	"testing"
)

const (
	mockAccount  = "testAccount"
	mockEmail    = "test@out.com"
	mockPassword = "password"

	mockBadAccount = "test"
)

const testMysqlConfigFilePath = "../default_config_files/default_mysql_config.json"

func TestInitMysqlConnection(t *testing.T) {
	if err := InitMysqlConnection(testMysqlConfigFilePath); err != nil {
		t.Fatalf("init mysql failed. reason: %s", err.Error())
	}
}

func TestRegister(t *testing.T) {
	if _, err := Register(mockAccount, mockEmail, mockPassword); err != nil {
		t.Fatalf("Mysql register user failed: %s", err.Error())
	}
}

func TestLogin(t *testing.T) {
	if id, err := Login(mockAccount, mockPassword); err != nil {
		t.Fatalf("Mysql login failed: %s", err.Error())
	} else if id == -1 {
		t.Fatalf("Mysql login user not found")
	}

	if id, err := Login(mockBadAccount, mockPassword); err != nil {
		t.Fatalf("Mysql login failed: %s", err.Error())
	} else if id != -1 {
		t.Fatalf("Mysql login wrong user found")
	}
}

func TestDrop(t *testing.T) {
	if err := dropTable(mysqlCfg.UserInfoTable); err != nil {
		t.Errorf("Mysql Drop table failed: %s", err.Error())
	}
}
