package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"QA-System/internal/global/config"
	"QA-System/internal/pkg/log"
)

var QA string
var Record string

func MongodbInit() *mongo.Database {
	// Get MongoDB connection information from the configuration file
	user := global.Config.GetString("mongodb.user")
	pass := global.Config.GetString("mongodb.pass")
	host := global.Config.GetString("mongodb.host")
	port := global.Config.GetString("mongodb.port")
	db := global.Config.GetString("mongodb.db")
	QA = global.Config.GetString("mongodb.qa-collection")
	Record = global.Config.GetString("mongodb.record-collection")

	// 构建 MongoDB 连接字符串
	dsn := fmt.Sprintf("mongodb://%v:%v@%v:%v/%v", user, pass, host, port, db)

	// 使用 dsn 连接 MongoDB
	clientOptions := options.Client().ApplyURI(dsn)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Logger.Fatal("Failed to connect to MongoDB:" + err.Error())
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Logger.Fatal("Failed to ping MongoDB:" + err.Error())
	}

	// Set the MongoDB database
	mdb := client.Database(db)

	// Print a log message to indicate successful connection to MongoDB
	log.Logger.Info("Connected to MongoDB")
	return mdb
}
