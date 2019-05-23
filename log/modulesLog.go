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
	enableModules map[string][]string
	mhState       *modulesHandlersState
)

// Environmental variable
const enableModuleLogKey = "ENABLE_MODULE_LOG"
const modulesLogKey = "MODULES_LOG"
const modulesDirKey = "MODULES_DIR"

type ModulesHandlerState interface {
	ModuleLogHandle(string, *Record)
}

func NewModulesHandlersState() *modulesHandlersState {
	mhs := &modulesHandlersState{state: make(map[string]Handler)}
	mhs.Init()
	return mhs
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
				if handler, ok := m.Get(module); ok {
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
	if os.Getenv(enableModuleLogKey) != "true" {
		return
	}
	if err := json.Unmarshal([]byte(os.Getenv(modulesLogKey)), &enableModules); err != nil {
		panic(fmt.Sprintf("read MODULES_LOG env error: %s", err))
	}
	if dir = os.Getenv(modulesDirKey); dir != "" {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("set MODULES_DIR env error: %s", err))
		}
	}
	ok = true
	return
}

func init() {
	enableModules = make(map[string][]string)
	mhState = NewModulesHandlersState()
}
