package core

import (
	"github.com/WAY29/FileNotifier/internal/parse"
	"github.com/WAY29/FileNotifier/internal/structs"
	"github.com/WAY29/FileNotifier/utils"
	"github.com/WAY29/errors"
)

var (
	Templates []*structs.Template
)

func LoadTemplates(templatePaths []string) {
	Templates = make([]*structs.Template, 0, len(templatePaths))

	var (
		template *structs.Template
		err      error
	)

	for _, path := range templatePaths {
		template, err = parse.ParseTemplate(path)
		if err != nil {
			nErr := errors.Wrapf(err, "Can't Parse template '%s'", path)
			utils.ErrorP(nErr)
		}
		Templates = append(Templates, template)
	}
}
