package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/WAY29/FileNotifier/utils"
	"github.com/WAY29/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var (
	Watcher    *fsnotify.Watcher
	DMP        *diffmatchpatch.DiffMatchPatch
	Events     string
	ContentMap sync.Map
)

func InitWatcher(files []string, directories []string, events []string) {
	utils.DebugF("Init %s", "watcher")

	var err error

	Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		utils.CliError("Create watcher error: "+err.Error(), 1)
	}

	DMP = diffmatchpatch.New()

	for _, path := range files {
		err = Watcher.Add(path)
		if err != nil {
			nErr := errors.Wrapf(err, "Can't watch file '%s'", path)
			utils.ErrorP(nErr)
		} else {
			addOrUpdateContentMap(path, true)
		}
	}

	// 递归监听文件夹
	for _, dirPath := range directories {
		watchDirectorys(dirPath)
	}

	Events = strings.Join(events, ",")
	Events = strings.ToLower(Events)

}

func watchDirectorys(dirPath string) {
	var err error

	err = Watcher.Add(dirPath)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't watch directory '%s'", dirPath)
		utils.ErrorP(nErr)
	}

	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't read directory '%s'", dirPath)
		utils.ErrorP(nErr)
	}

	PthSep := string(os.PathSeparator)

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			watchDirectorys(dirPath + PthSep + fi.Name())
		}
		addOrUpdateContentMap(dirPath+PthSep+fi.Name(), true)
	}
}

func addOrUpdateContentMap(path string, add bool) {
	var (
		absPath string
		content string
		err     error
	)

	absPath, err = utils.AbsFilePath(path)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't get file abspath '%s'", path)
		utils.ErrorP(nErr)
		return
	}

	content, err = utils.ReadFileAsString(absPath)
	ContentMap.Store(absPath, content)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't read file '%s'", absPath)
		utils.ErrorP(nErr)
		return
	}

	if add {
		utils.InfoF("Watch file/direcotry: '%s'", absPath)
	}
}

func Diff(path string) []diffmatchpatch.Diff {
	var (
		absPath    string
		oldContent interface{}
		content    string

		ok  bool
		err error
	)
	absPath, err = utils.AbsFilePath(path)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't get file abspath '%s'", path)
		utils.ErrorP(nErr)
		return nil
	}
	// 必须增加一个短暂的sleep，否则读取到的文件不准确
	time.Sleep(50 * time.Millisecond)

	content, err = utils.ReadFileAsString(absPath)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't read file '%s'", absPath)
		utils.ErrorP(nErr)
		return nil
	}

	if oldContent, ok = ContentMap.Load(absPath); !ok {
		return DMP.DiffMain("", content, true)
	} else {
		return DMP.DiffMain(oldContent.(string), content, true)
	}
}

func inEvents(event string) bool {
	return strings.Contains(Events, event)
}

func StartWatcher() {
	var (
		text    string
		absPath string

		modified = false
	)

	go func() {
		for {
			select {
			case event, ok := <-Watcher.Events:
				if !ok {
					continue
				}
				text = ""
				modified = false

				if event.Op&fsnotify.Write == fsnotify.Write && inEvents("write") {

					diffs := Diff(event.Name)
					if len(diffs) > 0 {
						for _, diff := range diffs {
							if diff.Type == diffmatchpatch.DiffEqual {
								continue
							}

							modified = true
							text += fmt.Sprintf("[%s] %s\n", diff.Type, strings.TrimSpace(diff.Text))
							utils.DebugF("Action:%s File:%s Text:%#v", diff.Type, event.Name, strings.TrimSpace(diff.Text))
						}
					}
					if modified {
						utils.SuccessF("Modified file: %s", event.Name)
					}
					addOrUpdateContentMap(event.Name, false)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename && inEvents("rename") {
					text = fmt.Sprintf("[Rename] %s", event.Name)
					utils.SuccessF("Rename file: %s", event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove && inEvents("remove") {
					text = fmt.Sprintf("[Remove] %s", event.Name)
					utils.SuccessF("Remove file: %s", event.Name)
				}
				text = strings.TrimSpace(text)
				absPath, _ = utils.AbsFilePath(event.Name)
				SendNotify(absPath, text)
			case err, ok := <-Watcher.Errors:
				if !ok {
					continue
				}
				nErr := errors.Wrap(err, "Watcher error")
				utils.ErrorP(nErr)
			}
		}

	}()
}
