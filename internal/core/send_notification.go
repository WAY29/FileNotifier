package core

import (
	"encoding/json"
	"net/url"
	"os/exec"
	"strings"

	"github.com/WAY29/FileNotifier/utils"
	"github.com/WAY29/errors"
	"github.com/go-resty/resty/v2"
)

var (
	Client = resty.New()
)

func render(raw, filename, text string) string {
	raw = strings.ReplaceAll(raw, "{{text}}", text)
	raw = strings.ReplaceAll(raw, "{{filename}}", filename)
	return raw
}

func vRender(raw, filename, text string) string {
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "\\r")
	filename = strings.ReplaceAll(filename, "\n", "\\n")
	filename = strings.ReplaceAll(filename, "\r", "\\r")

	raw = strings.ReplaceAll(raw, "{{text}}", text)
	raw = strings.ReplaceAll(raw, "{{filename}}", filename)
	return raw
}

func SendNotify(filename, text string) {
	var (
		c *exec.Cmd

		targetUrl, body string
		headers         map[string]string
		Encodedtext     string

		output    []byte
		err, nErr error
	)
	if text == "" {
		return
	}

	for _, template := range Templates {
		// 对windows路径进行处理
		filename = strings.ReplaceAll(filename, "\\", "/")

		for _, command := range template.TextCommandChain {
			command = vRender(command, filename, text)
			c, err = utils.ExecCommand(command)
			if err != nil {
				nErr = errors.Wrapf(err, "Exec command[%s] error", command)
				utils.ErrorP(nErr)
				return
			}
			output, err = c.Output()
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					utils.WarningF("Exec command[%s] return status %d, skip send notification", command, exitError.ExitCode())
				} else {
					nErr = errors.Wrapf(err, "Get commond[%s] output error", command)
					utils.ErrorP(nErr)
				}
				return
			}
			text = strings.ReplaceAll(string(output), "\\n", "\n")
			text = strings.ReplaceAll(text, "\\r", "\r")
			text = strings.TrimSpace(text)
			if text == "" {
				return
			}
		}
		for _, command := range template.FilenameCommandChain {
			command = vRender(command, filename, text)
			c, err = utils.ExecCommand(command)
			if err != nil {
				nErr = errors.Wrapf(err, "Exec command[%s] error", command)
				utils.ErrorP(nErr)
				return
			}
			output, err = c.Output()
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					utils.WarningF("Exec command[%s] return status %d, skip send notification", command, exitError.ExitCode())
				} else {
					nErr = errors.Wrapf(err, "Get commond[%s] output error", command)
					utils.ErrorP(nErr)
				}
				return
			}
			filename = strings.ReplaceAll(string(output), "\\n", "\n")
			filename = strings.ReplaceAll(filename, "\\r", "\r")
			filename = strings.TrimSpace(filename)
			if filename == "" {
				return
			}
		}

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

		utils.DebugF("debug: %s %#v %#v", body, text, filename)

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
