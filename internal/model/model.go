package model

func Init() {
	newRedis()
	newPostgres()
	newElasticsearch()
}
