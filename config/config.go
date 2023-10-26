package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/vrischmann/envconfig"
)

type configStruct struct {
	PORT                    string `envconfig:"default=5000"`
	SQL_DB_URL              string `envconfig:"default=root:@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"`
	JWT_KEY                 string `envconfig:"default=ToMaTo"`
	JWT_VALIDITY            int    `envconfig:"default=50"`
	SQL_MAX_OPEN_CONNECTION int    `envconfig:"default=20"`
	SQL_MAX_IDLE_CONNECTION int    `envconfig:"default=5"`
	REDIS_URL               string `envconfig:"default=127.0.0.6379"`
}

var ConfigEnv configStruct

func InitEnvVariables() {

	if err := godotenv.Load(".env"); err != nil {
		log.Println(err)
	}

	if err := envconfig.Init(&ConfigEnv); err != nil {
		log.Fatalln(err)
	} else {
		log.Println("Config Loaded Success")
	}
}
