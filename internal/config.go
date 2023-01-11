package internal

type RedisDbConfig struct {
	Server        string `envconfig:"REDIS_DB"`
	Password      string `envconfig:"REDIS_DB_PASSWORD"`
	RedisCa       string `envconfig:"REDIS_DB_CA"`
	RedisUserCert string `envconfig:"REDIS_DB_USER_CERT"`
	RedisUserKey  string `envconfig:"REDIS_DB_USER_KEY"`
	Index         string `envconfig:"REDIS_DB_INDEX"`
	MinIdle       int    `envconfig:"REDIS_DB_MINIDLE"`
	MaxActive     int    `envconfig:"REDIS_DB_MAXACTIVE"`
	IdleTimeout   int64  `envconfig:"REDIS_DB_IDLE_TIMEOUT"`
}

type RedisSearchConfig struct {
	Server        string `envconfig:"REDIS_SEARCH"`
	Password      string `envconfig:"REDIS_SEARCH_PASSWORD"`
	RedisCa       string `envconfig:"REDIS_SEARCH_CA"`
	RedisUserCert string `envconfig:"REDIS_SEARCH_USER_CERT"`
	RedisUserKey  string `envconfig:"REDIS_SEARCH_USER_KEY"`
	Index         string `envconfig:"REDIS_SEARCH_INDEX"`
	MinIdle       int    `envconfig:"REDIS_SEARCH_MINIDLE"`
	MaxActive     int    `envconfig:"REDIS_SEARCH_MAXACTIVE"`
	IdleTimeout   int64  `envconfig:"REDIS_SEARCH_IDLE_TIMEOUT"`
}

type Config struct {
	RedisDbConfig     RedisDbConfig
	RedisSearchConfig RedisSearchConfig
	RepoImpl          string `envconfig:"REPO_IMPL" required:"true"`
	BrokerImpl        string `envconfig:"BROKER_IMPL" required:"true"`
	GrpcHost          string `envconfig:"GRPC_HOST" required:"true"`
	UseRedisSearch    string `envconfig:"USE_REDIS_SEARCH"`
	ServiceGrpcPort   string `envconfig:"SERVICE_PORT" required:"true"`
}
