package log

import (
	"sync"
	"fmt"
	"strings"
	"os"
	"path/filepath"
	"encoding/json"
)

var (
	enableModules   = make(map[string][]string)
	mhState         *modulesHandlersState
	moduleLogLvl    = LvlTrace
	moduleParamsStr string
)

type ModulesHandlerState interface {
	ModuleLogHandle(string, *Record)
}

func SetModuleLogLvl(lvl Lvl) {
	moduleLogLvl = lvl
}

func SetModuleParamsStr(str string) {
	moduleParamsStr = str
}

func InitModulesHandlersState() {
	mhState.Init()
}

func newModulesHandlersState() *modulesHandlersState {
	return &modulesHandlersState{state: make(map[string]Handler)}
}

// modules log will be written to the specified dir
type modulesHandlersState struct {
	state map[string]Handler
	mu    sync.RWMutex
	once  sync.Once
	ModulesHandlerState
	dir   string
}

func (m *modulesHandlersState) Init() {
	m.once.Do(func() {
		m.init()
	})
}

func (m *modulesHandlersState) Put(k string, h Handler) {
	defer m.mu.Unlock()
	m.mu.Lock()
	m.state[k] = h
}

func (m *modulesHandlersState) Get(k string) (Handler, bool) {
	defer m.mu.RUnlock()
	m.mu.RLock()
	h, ok := m.state[k]
	return h, ok
}

func (m *modulesHandlersState) GetStateFilePath(k string) string {
	return filepath.Join(m.dir, fmt.Sprintf("%s.log", k))
}

// different pkg correspond to different modules
func (m *modulesHandlersState) ModuleLogHandle(pkg string, record *Record) {
	if len(enableModules) == 0 {
		return
	}
	for module, vList := range enableModules {
		for _, v := range vList {
			if strings.Contains(pkg, v) {
				if handler, ok := m.Get(module); ok && record.Lvl <= moduleLogLvl {
					handler.Log(record)
				}
				return
			}
		}
	}
}

func (m *modulesHandlersState) init() {
	dir, ok := EnableModulesLog()
	if !ok {
		return
	}
	m.dir = dir
	for k := range enableModules {
		if _, ok := m.Get(k); !ok {
			m.Put(k, must(FileHandler(m.GetStateFilePath(k), JSONFormat())))
		}
	}
}

// get the specified module based on the environment variable
func EnableModulesLog() (dir string, ok bool) {
	if moduleParamsStr == "" {
		return
	}
	if err := json.Unmarshal([]byte(moduleParamsStr), &enableModules); err != nil {
		panic(fmt.Sprintf("read modules_log_json_params error: %s", err))
	}

	if dirL, ok := enableModules["__dir__"]; ok && len(dirL) != 0 && dirL[0] != "" {
		dir = dirL[0]
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("create modules log dir error: %s", err))
		}
		delete(enableModules, "__dir__")
	}
	ok = true
	return
}

func init() {
	mhState = newModulesHandlersState()
}
