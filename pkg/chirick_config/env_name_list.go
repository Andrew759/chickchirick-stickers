package chirick_config

// server
const (
	ServerUrl   = "SERVER_URL"
	UserApp     = "USER_APP"
	AuthApp     = "AUTH_APP"
	FrontendApp = "FRONTEND_URL"
	MessagesApp = "MESSAGES_APP_URL"
	AdminToken  = "STICKERS_ADMIN_TOKEN"
)

// environment
const (
	Enviroment = "ENVIRONMENT"
)

// database
const (
	DbHost         = "DB_HOST"
	DbPort         = "DB_PORT"
	DbInternalPort = "DB_INTERNAL_PORT"
	DbName         = "DB_NAME"
	DbUser         = "DB_USER"
	DbPass         = "DB_PASS"
	DbTimezone     = "DB_TIMEZONE"
)

// redis
const (
	RedisHost         = "REDIS_HOST"
	RedisPort         = "REDIS_PORT"
	RedisInternalPort = "REDIS_INTERNAL_PORT"
	RedisDB           = "REDIS_DB"
	RedisUser         = "REDIS_USER"
	RedisPassword     = "REDIS_PASSWORD"
)

const (
	AppPort    = "APP_PORT"
	GinMode    = "GIN_MODE"
	DBHost     = "DB_HOST"
	DBPort     = "DB_PORT"
	DBSSLMode  = "DB_SSLMODE"
	DBTimezone = "DB_TIMEZONE"
)
