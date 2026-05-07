package handlers

var (
	DiscordBotConfigValues BotConfig
)

type AuthConfig struct {
	Token string `yaml:"token"`
}

type StarConfig struct {
	Threshold int `yaml:"threshold"`
}

type SettingsConfig struct {
	Star StarConfig `yaml:"star"`
}

type DiscordConfig struct {
	CommandPrefix             string      `yaml:"cmd_prefix"`
	GuildID                   string      `yaml:"guildID"`
	Auth                      AuthConfig  `yaml:"auth"`
	Channels                  []string    `yaml:"threads"`
	AllowedURLPrefix          []string    `yaml:"allowed-urls-threads-filtered"`
	StarBoardChannel          string      `yaml:"starboard-channel"`
	AnnounceChannel           string      `yaml:"announcement-channel"`
	ConfessionChannel         string      `yaml:"confession-channel"`
	ModLogChannel             string      `yaml:"mod-log-channel"`
	VentChannel               string      `yaml:"vent-channel"`
	ConfessionCooldown        int         `yaml:"confession-cooldown"`
	ModChannel                string      `yaml:"mod-channel"`
	StarBoardFilteredChannels []string    `yaml:"starboard-disabled-channels"`
	Redis                     RedisConfig `yaml:"redis"`
	HelloReply                []string    `yaml:"hello-replies"`
}

type RedisConfig struct {
	Endpoint string `yaml:"endpoint"`
	Password string `yaml:"password"`
	DbNum    int    `yaml:"dbnum"`
	DoneVal  int    `yaml:"doneval"`
}

type BotConfig struct {
	SettingsConfig `yaml:"settings"`
	DiscordConfig  `yaml:"discord"`
	RedisConfig    `yaml:"redis"`
	Version        string `yaml:"version"`
}

type Handler struct {
	Name     string
	Function interface{}
	File     string
}
