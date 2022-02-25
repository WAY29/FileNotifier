package structs

type Template struct {
	Name          string            `yaml:"name"`
	Url           string            `yaml:"url"`
	Method        string            `yaml:"method"`
	Headers       map[string]string `yaml:"headers"`
	Body          string            `yaml:"body"`
	UrlEncodeText bool              `yaml:"urlencode_text"`
	EscapeJson    bool              `yaml:"esacpe_json"`
}
