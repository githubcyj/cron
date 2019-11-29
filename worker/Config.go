package worker

import (
	"encoding/json"
	"flag"
	"io/ioutil"
)

//配置文件
type Config struct {
	EtcdIP          []string `json:"etcdIP"`
	EtcdTimeout     int      `json:"etcdTimeout"`
	LogFilename     string   `json:"logFilename"`
	LogFilepath     string   `json:"logFilepath"`
	AutoCommitTime  int      `json:"autoCommitTime"`
	AutoCommitCount int      `json:"autoCommitCount"`
	JobRuntime      int      `json:"jobRuntime"`
	Dialect         string   `json:"dialect"`
	Database        string   `json:"database"`
	User            string   `json:"user"`
	Password        string   `json:"password"`
	Charset         string   `json:"charset"`
	Host            string   `json:"host"`
	Port            int      `json:"port"`
	MaxIdleConns    int      `json:"maxIdleConns"`
	MaxOpenConns    int      `json:"maxOpenConns"`
	RedisHost       string   `json:"redisHost"`
	RedisPort       int      `json:"redisPort"`
	RedisPassword   string   `json:"redisPassword"`
	RedisMaxIdle    int      `json:"redisMaxIdle"`
	RedisMaxActive  int      `json:"redisMaxActive"`
}

var (
	GConfig *Config
)

func InitConfig() (err error) {
	var (
		bytes    []byte
		conf     Config
		confFile string
	)

	//读取master.json,可让用户自定义
	//如 master -config ./master.json 默认master.json
	flag.StringVar(&confFile, "config", "./worker.json", "指定配置文件")
	//解析命令行参数
	flag.Parse()

	//读取文件
	if bytes, err = ioutil.ReadFile(confFile); err != nil {
		return
	}

	//反序列化
	if err = json.Unmarshal(bytes, &conf); err != nil {
		return
	}

	GConfig = &conf

	return
}
