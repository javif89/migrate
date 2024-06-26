package database

type Config struct {
	Username string
	Password string
	Host string
	Port string
	Database string
}

type DriverName string

// Driver names for convenience
var DriverMysql DriverName = "mysql"
var DriverSqlite DriverName = "sqlite"

var Drivers map[DriverName]Driver = map[DriverName]Driver {
	DriverMysql: MysqlDriver{},
	DriverSqlite: SQLiteDriver{},
}

func GetDriver(driver DriverName, cfg Config) (Driver, error) {
	d := Drivers[driver]

	return d.Open(cfg)
}