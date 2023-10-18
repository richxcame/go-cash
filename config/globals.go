package config

import (
	"log"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

type Globals struct {
	JWT_SECRET            []byte
	ACCESS_TOKEN_TIMEOUT  int
	REFRESH_TOKEN_TIMEOUT int
}

var GlobalConfig Globals

func InitGlobals() {
	GlobalConfig.JWT_SECRET = []byte(os.Getenv("JWT_SECRET"))
	accessTokenTimeout := os.Getenv("ACCESS_TOKEN_TIMEOUT")
	accessTimeout, err := strconv.Atoi(accessTokenTimeout)
	if err != nil {
		log.Fatalf("couldn't convert access token timeout to integer: %v", err)
	}
	GlobalConfig.ACCESS_TOKEN_TIMEOUT = accessTimeout

	refreshTokenTimeout := os.Getenv("REFRESH_TOKEN_TIMEOUT")
	refreshTimeout, err := strconv.Atoi(refreshTokenTimeout)
	if err != nil {
		log.Fatalf("couldn't convert refresh token timeout to integer: %v", err)
	}
	GlobalConfig.REFRESH_TOKEN_TIMEOUT = refreshTimeout
}
