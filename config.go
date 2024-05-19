package mysql

type Config struct {
	Addr     string `mapstructure:"address" json:"address" yaml:"address"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DataBase string `mapstructure:"database" json:"database" yaml:"database"`
}
