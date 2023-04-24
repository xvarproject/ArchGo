package openai_config

var CLI struct {
	Verbose bool   `help:"Verbose mode."`
	Config  string `help:"Config file." name:"openai-config" type:"file" default:"openai-config.json"`
}
