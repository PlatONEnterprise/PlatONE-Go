package log

import (
	"sync"
	"fmt"
	"strings"
	"path/filepath"
	"encoding/json"
	"strconv"
	"os"
)

var (
	enableModules = make(map[string][]string)
	moduleLogLvl  = LvlTrace
)

var (
	mhState         *modulesHandlersState
	vModule         string
	backtraceAt     string
	moduleParamsStr string
)

const (
	moduleDirKey          = "__dir__"
	moduleFileSizeKey     = "__size__"
	moduleDefaultFileSize = 262144
)

type ModulesHandlerState interface {
	ModuleLogHandle(string, *Record)
}

func SetVModule(v string) {
	vModule = v
}

func SetBacktraceAt(b string) {
	backtraceAt = b
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
	size  uint
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
	return filepath.Join(m.dir, k)
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
	dir, ok, size := EnableModulesLog()
	if !ok {
		return
	}
	m.dir, m.size = dir, size
	for k := range enableModules {
		if _, ok := m.Get(k); !ok {
			g := NewGlogHandler(must(RotatingFileHandler(
				m.GetStateFilePath(k),
				m.size,
				TerminalFormat(true)),
			))
			g.Verbosity(moduleLogLvl)
			g.Vmodule(vModule)
			g.BacktraceAt(backtraceAt)
			m.Put(k, g)
		}
	}
}

// Start the module log and set the configuration
func EnableModulesLog() (dir string, ok bool, size uint) {
	if moduleParamsStr == "" {
		return
	}
	if err := json.Unmarshal([]byte(moduleParamsStr), &enableModules); err != nil {
		panic(fmt.Sprintf("read modules_log_json_params error: %s", err))
	}

	if d, ok := getEnableModulesConfigKey(moduleDirKey, func(s string) interface{} {
		if err := os.MkdirAll(s, os.ModePerm); err != nil {
			panic(fmt.Sprintf("create modules log dir error: %s", err))
		}
		return s
	}).(string); ok {
		dir = d
	}

	if s, ok := getEnableModulesConfigKey(moduleFileSizeKey, func(sizeStr string) interface{} {
		var (
			s   int
			err error
		)
		if s, err = strconv.Atoi(sizeStr); err != nil {
			panic(fmt.Sprintf("strconv.Atoi error; check the __size__: %s", err))
		}
		return s
	}).(int); ok {
		size = uint(s)
	}

	if size == 0 {
		size = moduleDefaultFileSize
	}
	ok = true
	return
}

func getEnableModulesConfigKey(key string, fn func(string) interface{}) interface{} {
	if v, ok := enableModules[key]; ok {
		delete(enableModules, key)
		if len(v) != 0 && v[0] != "" {
			return fn(v[0])
		}
	}
	return nil
}

func init() {
	mhState = newModulesHandlersState()
}
