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

	WatchNumber = 0
)

func InitWatcher(files, directories, excludes, events []string) int {
	var (
		err     error
		absPath string
	)

	utils.DebugF("Init %s", "watcher")

	Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		utils.CliError("Create watcher error: "+err.Error(), 1)
	}

	DMP = diffmatchpatch.New()

	// 拼接events
	Events = strings.Join(events, ",")
	Events = strings.ToLower(Events)

	// 处理excludes
	for n, path := range excludes {
		absPath, err = utils.AbsFilePath(path)
		if err != nil {
			nErr := errors.Wrapf(err, "Can't get abspath %#v", path)
			utils.ErrorP(nErr)
		} else {
			excludes[n] = absPath
		}
	}

	// 递归监听文件
	for _, path := range files {
		addWatch(path, excludes)
	}
	// 递归监听文件夹
	for _, path := range directories {
		addWatch(path, excludes)
	}

	return WatchNumber
}

func addWatch(path string, excludes []string) {
	var (
		err     error
		absPath string
	)

	absPath, err = utils.AbsFilePath(path)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't get abspath %#v", path)
		utils.ErrorP(nErr)
		return
	}

	for _, excludePath := range excludes {
		if strings.HasPrefix(absPath, excludePath) {
			return
		}
	}

	err = Watcher.Add(absPath)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't watch %#v", absPath)
		utils.ErrorP(nErr)
	} else {
		utils.InfoF("Watch file/direcotry: %#v", absPath)
		WatchNumber += 1
	}

	if utils.IsDir(absPath) {
		dir, err := ioutil.ReadDir(absPath)
		if err != nil {
			nErr := errors.Wrapf(err, "Can't read directory %#v", absPath)
			utils.ErrorP(nErr)
		}

		PthSep := string(os.PathSeparator)

		for _, fi := range dir {
			addWatch(absPath+PthSep+fi.Name(), excludes)
			addOrUpdateContentMap(absPath + PthSep + fi.Name())
		}
	}
}

func addOrUpdateContentMap(path string) {
	var (
		absPath string
		content string
		err     error
	)

	absPath, err = utils.AbsFilePath(path)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't get file abspath %#v", path)
		utils.ErrorP(nErr)
		return
	}

	if !utils.IsFile(absPath) {
		return
	}

	content, err = utils.ReadFileAsString(absPath)
	ContentMap.Store(absPath, content)
	if err != nil {
		nErr := errors.Wrapf(err, "Can't read file %#v", absPath)
		utils.ErrorP(nErr)
		return
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
					addOrUpdateContentMap(event.Name)
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
