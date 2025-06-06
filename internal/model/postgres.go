package model

import (
	"BBingyan/internal/config"
	"BBingyan/internal/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func newPostgres() {
	var db *gorm.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(config.Config.Postgres.Dsn), &gorm.Config{})
		if err != nil {
			log.Warnf("%d Fail to connect to database", i)
		} else {
			break
		}
	}
	if err != nil {
		log.Fatalf("Fail to connect to database")
	}
	DB = db

	//AutoMigrate
	if err := DB.AutoMigrate(&User{}); err != nil {
		log.Fatalf("Fail to automigrate database")
	}
	if err := DB.AutoMigrate(&Post{}); err != nil {
		log.Fatalf("Fail to automigrate database")
	}
	if err := DB.AutoMigrate(&Comment{}); err != nil {
		log.Fatalf("Fail to automigrate database")
	}
	if err := DB.AutoMigrate(&FollowShip{}); err != nil {
		log.Fatalf("Fail to automigrate database")
	}
	if err := DB.AutoMigrate(&UserLikeShip{}); err != nil {
		log.Fatalf("Fail to automigrate database")
	}
	if err := DB.AutoMigrate(&PostLikeShip{}); err != nil {
		log.Fatalf("Fail to automigrate database")
	}
	if err := DB.AutoMigrate(&CommentLikeShip{}); err != nil {
		log.Fatalf("Fail to automigrate database")
	}

	m := DB.Migrator()
	pwd, err := bcrypt.GenerateFromPassword([]byte(config.Config.Admin.Password), 12)
	if err != nil {
		log.Fatalf("Fail to hash admin pwd,err:%s", err)
	}
	admin := &Admin{
		Name:       config.Config.Admin.Name,
		Password:   string(pwd),
		AddedAdmin: config.Config.Admin.Name,
	}
	if !m.HasTable(&Admin{}) {
		if err := DB.AutoMigrate(&Admin{}); err != nil {
			log.Fatalf("Fail to automigrate database")
		}
		if err := DB.Model(&Admin{}).Create(admin).Error; err != nil {
			log.Fatalf("Fail to init admin table,err:%v", err)
		}
	} else {
		DB.Model(&Admin{}).Create(admin)
	}
	if !m.HasTable(&Tag{}) {
		if err := DB.AutoMigrate(&Tag{}); err != nil {
			log.Fatalf("Fail to automigrate database")
		}
		for _, tag := range config.Config.Curd.Tags {
			tagNode := &Tag{
				Tag:        tag,
				AddedAdmin: config.Config.Admin.Name,
			}
			if err := DB.Model(&Tag{}).Create(tagNode).Error; err != nil {
				log.Fatalf("Fail to init tag table,err:%v", err)
			}
		}
	}

	log.Infof("Finish initializing postgres")
}
