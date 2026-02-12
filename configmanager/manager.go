package configmanager

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

/*
## Конфигуратор приложения с `sync.Once`

**Описание**
Этот проект реализует **потокобезопасный** менеджер конфигурации, который загружает настройки **только один раз** при первом запросе.
Используется `sync.Once`, чтобы избежать повторной загрузки при одновременном доступе из нескольких горутин.
---

**Возможности**
1. Ленивая инициализация – загрузка конфигурации только при первом вызове.
2. Потокобезопасность – отсутствие гонок данных при многопоточной работе.
3. Гибкость – возможность загружать конфигурацию из файла, переменных окружения или базы данных.
---

**Реализованные методы**
- `LoadConfig()` – загружает конфигурацию **один раз** и сохраняет в памяти.
- `Get(key string) string` – возвращает значение конфигурации по ключу.
- `PrintConfig()` – выводит загруженные параметры.
---


*/

var (
	ErrEmptyLoaders  = errors.New("empty config. loaders dont given")
	ErrAlreadyLoaded = errors.New("config already loaded")
)

type Config map[string]string
type Loader func() (Config, error)

type IConfigManager interface {
	LoadConfig() error
	Get(key string) string
	PrintConfig()
}

type ConfigManager struct {
	config Config

	loaders []Loader
	once    *sync.Once
	loaded  atomic.Bool

	loadErr error

	mu sync.Mutex
}

func NewConfigManager(loaders ...Loader) *ConfigManager {
	return &ConfigManager{
		config:  Config{},
		loaders: loaders,
		once:    &sync.Once{},

		loaded:  atomic.Bool{},
		loadErr: nil,

		mu: sync.Mutex{},
	}
}

func (cm *ConfigManager) AddLoader(loader Loader) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.loaded.Load() {
		return ErrAlreadyLoaded
	}

	cm.loaders = append(cm.loaders, loader)

	return nil
}

func (cm *ConfigManager) LoadConfig() error {
	cm.once.Do(func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()

		cm.loadErr = cm.load()
		if cm.loadErr == nil {
			cm.loaded.Store(true)
		}
	})

	return cm.loadErr
}

func (cm *ConfigManager) Get(key string) string {
	if !cm.loaded.Load() {
		return ""
	}

	return cm.config[key]
}

func (cm *ConfigManager) PrintConfig() {
	if !cm.loaded.Load() {
		fmt.Println("not loaded")
		return
	}

	for k, v := range cm.config {
		fmt.Printf("%s:%s\n", k, v)
	}
}

func (cm *ConfigManager) load() error {
	if len(cm.loaders) == 0 {
		return ErrEmptyLoaders
	}

	buf := make(chan Config, len(cm.loaders))
	errBuf := make(chan error, len(cm.loaders))

	wg := sync.WaitGroup{}
	for _, loader := range cm.loaders {
		wg.Add(1)
		go func(l Loader) {
			defer wg.Done()

			cnf, err := l()

			buf <- cnf
			errBuf <- err
		}(loader)
	}

	wg.Wait()
	close(buf)
	close(errBuf)

	errs := make([]error, 0, len(cm.loaders))
	for e := range errBuf {
		if e != nil {
			errs = append(errs, e)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	for c := range buf {
		cm.save(c)
	}

	return nil
}

func (cm *ConfigManager) save(cnf Config) {
	for k, v := range cnf {
		cm.config[k] = v
	}
}
