package db

import (
	"context"
	"database/sql"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func StartDbTransaction(fn func(mysqlTx *gorm.DB, mongoTx mongo.SessionContext) error) error {
	err := error(nil)
	mongoSess, err := mongoConnection.StartSession()
	if err != nil {
		return err
	}
	defer mongoSess.EndSession(context.Background())

	mysqlTx := mysqlDb.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if mysqlTx.Error != nil {
		return mysqlTx.Error
	}

	mongoTx := mongo.NewSessionContext(context.Background(), mongoSess)
	err = mongoTx.StartTransaction()
	if err != nil {
		mysqlTx.Rollback()
		return err
	}

	err = fn(mysqlTx, mongoTx)
	if err != nil {
		mysqlTx.Rollback()
		mongoTx.AbortTransaction(context.Background())
	} else {
		mysqlTx.Commit()
		mongoTx.CommitTransaction(context.Background())
	}

	return err
}
