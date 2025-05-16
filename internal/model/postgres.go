package model

import (
	"BBingyan/internal/config"
	"BBingyan/internal/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func newPostgres() {
	db, err := gorm.Open(postgres.Open(config.Config.Postgres.Dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Fail to connect to postgres")
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
	if !m.HasTable(&Admin{}) {
		if err := DB.AutoMigrate(&Admin{}); err != nil {
			log.Fatalf("Fail to automigrate database")
		}
		admin := &Admin{
			Name:       config.Config.Admin.Name,
			Password:   config.Config.Admin.Password,
			AddedAdmin: config.Config.Admin.Name,
		}
		if err := DB.Model(&Admin{}).Create(admin).Error; err != nil {
			log.Fatalf("Fail to init admin table,err:%v", err)
		}
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
