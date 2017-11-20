package consts

const (
	CONFIG_FILE                   = "../config/shkaff.json"
	DEFAULT_HOST                  = "localhost"
	DEFAULT_DATABASE_PORT         = 5432
	DEFAULT_AMQP_PORT             = 5672
	DEFAULT_DATABASE_DB           = "postgres"
	DEFAULT_REFRESH_DATABASE_SCAN = 60

	INVALID_DATABASE_HOST     = "Database host in config file is empty. Shkaff set '%s'\n"
	INVALID_DATABASE_PORT     = "Database port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_DATABASE_DB       = "Database name in config file is empty. Shkaff set '%s'\n"
	INVALID_DATABASE_USER     = "Database user name is empty"
	INVALID_DATABASE_PASSWORD = "Database password is empty"

	INVALID_AMQP_HOST     = "AMQP host in config file is empty. Shkaff set '%s'\n"
	INVALID_AMQP_PORT     = "AMPQ port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_AMQP_USER     = "AMQP user name is empty"
	INVALID_AMQP_PASSWORD = "AMQP password is empty"

	RMQ_URI_TEMPLATE  = "amqp://%s:%s@%s:%d/%s"
	PSQL_URI_TEMPLATE = "postgres://%s:%s@%s:%d/%s?sslmode=disable"

	REQUEST_GET_STARTTIME = "SELECT task_id, start_time, verb, thread_count, ipv6, gzip, host, port, databases, db_user, db_password,tp.\"type\" as db_type FROM shkaff.tasks t INNER JOIN shkaff.db_settings db ON t.db_settings_id = db.db_id INNER JOIN shkaff.types tp ON tp.type_id = db.\"type\" WHERE t.start_time <= to_timestamp(%d) AND t.is_active = true;"
	REQUESR_UPDATE_ACTIVE = "UPDATE shkaff.tasks SET is_active = $1 WHERE task_id = $2;"

	CACHEPATH = "./cache/cache.db"
)
