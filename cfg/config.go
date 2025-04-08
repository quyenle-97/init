package cfg

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/joho/godotenv"
)

type DB struct {
	DBDriver string `json:"DB_DRIVER"`
	DBHost   string `json:"DB_HOST"`
	DBPort   string `json:"DB_PORT"`
	DBUser   string `json:"DB_USER"`
	DBPass   string `json:"DB_PASS"`
	DBName   string `json:"DB_NAME"`
}

type RConfig struct {
	Host    string   `json:"REDIS_HOST"` // redis host
	Port    string   `json:"REDIS_PORT"`
	Pass    string   `json:"REDIS_PASS"`    // redis pass
	Index   string   `json:"REDIS_INDEX"`   // redis index
	Addr    []string `json:"REDIS_ADDR"`    // redis addr
	Cluster string   `json:"REDIS_CLUSTER"` // redis cluster
}

func (r RConfig) RPort() int {
	cp, err := strconv.Atoi(r.Port)
	if err != nil {
		panic(err)
	}
	return cp
}

func (r RConfig) RIndex() int {
	cp, err := strconv.Atoi(r.Index)
	if err != nil {
		return 0
	}
	return cp
}

func (r RConfig) RCluster() bool {
	if r.Cluster == "true" {
		return true
	}
	return false
}

type Config struct {
	AppEnv   string `json:"APP_ENV"`
	BasePath string `json:"BASE_PATH"`
	DB
	RConfig
	Server
}

type Server struct {
	Port string `json:"SERVER_PORT"`
}

func (s Server) GetPort() int {
	port, err := strconv.Atoi(s.Port)
	if err != nil {
		return 80
	}
	return port
}

func LoadConfig() Config {
	var config Config
	data, err := godotenv.Read()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(jsonStr, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
