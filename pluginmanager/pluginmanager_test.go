package pluginmanager

import (
	"fmt"
	"testing"
	"time"
)

type TestPlugin struct{}

func (p *TestPlugin) Execute() string {
	return "TestPlugin executed successfully!"
}

func initDemo() (Plugin, error) {
	// Имитация длительной инициализации
	time.Sleep(500 * time.Millisecond)
	return &TestPlugin{}, nil
}

func TestPluginManager_Work(t *testing.T) {
	pm := NewPluginManager()

	plugName := "demo"

	_, err := pm.GetPlugin(plugName)
	if err == nil {
		t.Fatal("expect error not found plugin, got nil")
	}

	pm.RegisterPlugin(plugName, initDemo)

	err = pm.RegisterPlugin(plugName, initDemo)
	if err == nil {
		t.Fatal("expect error plugin name already taken, got nil")
	}

	p, err := pm.GetPlugin(plugName)
	if err != nil {
		t.Fatal("expect nil error when get first init plugin, got err ", err)
	}

	p, err = pm.GetPlugin(plugName)
	if err != nil {
		t.Fatal("expect nil error when get cached plugin, got err ", err)
	}

	s := p.Execute()
	fmt.Println(s)
}
