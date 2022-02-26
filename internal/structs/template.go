package structs

type Template struct {
	Name                 string            `yaml:"name"`
	Url                  string            `yaml:"url"`
	Method               string            `yaml:"method"`
	Headers              map[string]string `yaml:"headers"`
	Body                 string            `yaml:"body"`
	TextCommandChain     []string          `yaml:"text_command_chain"`
	FilenameCommandChain []string          `yaml:"filename_command_chain"`
	UrlEncodeText        bool              `yaml:"urlencode_text"`
	EscapeJson           bool              `yaml:"escape_json"`
}
