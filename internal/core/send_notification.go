package core

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/WAY29/FileNotifier/utils"
	"github.com/WAY29/errors"
	"github.com/go-resty/resty/v2"
)

var (
	Client = resty.New()
)

func render(raw, filename, text string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	raw = strings.ReplaceAll(raw, "{{text}}", text)
	raw = strings.ReplaceAll(raw, "{{filename}}", filename)
	return raw
}

func SendNotify(filename, text string) {
	var (
		targetUrl, body string
		headers         map[string]string

		Encodedtext string
	)
	if text == "" {
		return
	}

	for _, template := range Templates {
		if template.UrlEncodeText {
			Encodedtext = url.QueryEscape(text)
		} else {
			Encodedtext = text
		}

		if template.EscapeJson {
			tmpBytes, _ := json.Marshal(text)
			Encodedtext = string(tmpBytes)
			Encodedtext = Encodedtext[1 : len(Encodedtext)-1]
		} else {
			Encodedtext = strings.ReplaceAll(text, "\\", "\\\\")
		}

		targetUrl = render(template.Url, filename, url.PathEscape(text))
		headers = make(map[string]string, len(template.Headers))
		for key, value := range template.Headers {
			headers[key] = value
		}

		for k, v := range headers {
			headers[k] = render(v, filename, Encodedtext)
		}

		body = render(template.Body, filename, Encodedtext)
		utils.DebugF("%#v %#v %s", filename, Encodedtext, body)

		resp, err := Client.R().SetHeaders(headers).SetBody(body).Execute(template.Method, targetUrl)
		if err != nil {
			nErr := errors.Wrapf(err, "Template[%s] send text '%s' error", template.Name, text)
			utils.ErrorP(nErr)
			return
		}

		if resp.StatusCode() != 200 {
			utils.WarningF("Template[%s] StatusCode != 200", template.Name)
		}

		utils.DebugF("StatusCode:%d Response: %s", resp.StatusCode(), resp.String())
	}
}
