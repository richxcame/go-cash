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
	BOOKINGS_API_URL      string
	GOTOLEG_URL           string
	GOTOLEG_LOGIN         string
	GOTOLEG_PASS          string
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

	GlobalConfig.BOOKINGS_API_URL = os.Getenv("BOOKINGS_API_URL")
	GlobalConfig.GOTOLEG_URL = os.Getenv("GOTOLEG_URL")
	GlobalConfig.GOTOLEG_LOGIN = os.Getenv("GOTOLEG_LOGIN")
	GlobalConfig.GOTOLEG_PASS = os.Getenv("GOTOLEG_PASS")
}
