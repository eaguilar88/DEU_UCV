package entities

import (
	"log"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SimulateSignUpInterval    int
	SimulateOrbStatusInterval int
	SimulateSignUpPath        string
	SimulateOrbStatusPath     string
	TestFilePath              string
	Port                      int
}

func NewConfig() *Config {
	conf, err := godotenv.Read()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	signUpInterval, err := strconv.Atoi(conf["SIGNUP_INTERVAL"])
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	statusInterval, err := strconv.Atoi(conf["ORB_STATUS_INTERVAL"])
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port, err := strconv.Atoi(conf["PORT"])
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		Port:                      port,
		SimulateSignUpInterval:    signUpInterval,
		SimulateOrbStatusInterval: statusInterval,
		SimulateSignUpPath:        conf["SIGNUP_PATH"],
		SimulateOrbStatusPath:     conf["ORB_STATUS_PATH"],
		TestFilePath:              conf["TEST_FILE_PATH"],
	}
}
