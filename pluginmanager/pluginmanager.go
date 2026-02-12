package pluginmanager

import (
	"errors"
	"sync"
)

/*

Инициализация плагинов с `sync.Once`

**Цель задания**
Реализовать систему безопасной инициализации плагинов, где:
- Каждый плагин инициализируется **только один раз**
- Инициализация потокобезопасна
- Ошибки при инициализации корректно обрабатываются
- Плагины доступны для использования из разных компонентов
---

**Требования**
1. **Структура `PluginManager`**:
    - Хранит загруженные плагины
    - Использует `sync.Once` для каждого плагина
    - Поддерживает конкурентный доступ

2. **Методы**:
    - `GetPlugin(name string) (Plugin, error)`  возвращает инициализированный плагин
    - `RegisterPlugin()`  регистрирует плагины (симуляция)

*/

var (
	ErrPluginNotFound  = errors.New("plugin not found in registered")
	ErrPluginNameTaken = errors.New("plugin name already taken")
)

// Интерфейс для всех плагинов
type Plugin interface {
	Execute() string
}

// Управляет инициализацией и доступом к плагинам
type PluginManager struct {
	plugins map[string]*pluginEntry
	mu      sync.RWMutex
}

type pluginEntry struct {
	//Добавить необходимые поля для однократной инициализации
	plugin Plugin

	once    sync.Once
	initErr error
	initFn  func() (Plugin, error)
}

// NewPluginManager создает новый менеджер плагинов
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*pluginEntry),
		mu:      sync.RWMutex{},
	}
}

// RegisterPlugin регистрирует новый плагин
func (pm *PluginManager) RegisterPlugin(name string, initFn func() (Plugin, error)) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, found := pm.plugins[name]; found {
		return ErrPluginNameTaken
	}

	pEntry := &pluginEntry{
		once: sync.Once{},
	}

	pEntry.initFn = func() (Plugin, error) {
		pEntry.once.Do(func() {
			pEntry.plugin, pEntry.initErr = initFn()
		})

		return pEntry.plugin, pEntry.initErr
	}

	pm.plugins[name] = pEntry

	return nil
}

// GetPlugin возвращает инициализированный плагин
func (pm *PluginManager) GetPlugin(name string) (Plugin, error) {
	// Реализовать:
	// 1. Проверку существования плагина
	// 2. Потокобезопасную однократную инициализацию
	// 3. Обработку и кэширование ошибок
	// 4. Возврат кэшированного результата
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	entry, found := pm.plugins[name]
	if !found {
		return nil, ErrPluginNotFound
	}

	// плагин и ошибка кешируются внутри pluginEntry в initFn --> в once
	return entry.initFn()
}
