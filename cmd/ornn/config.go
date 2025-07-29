package ornn

type Config struct {
	DbType      string `json:"db_type"`
	Dsn         string `json:"dsn"`
	SchemaPath  string `json:"schema_path"`
	ConfigPath  string `json:"config_path"`
	GenPath     string `json:"gen_path"`
	FileName    string `json:"file_name"`
	PackageName string `json:"package_name"`
	ClassName   string `json:"class_name"`
}
