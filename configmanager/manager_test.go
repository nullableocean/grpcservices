package configmanager

import (
	"os"
	"sync"
	"testing"
)

func TestConfigManager_Work(t *testing.T) {
	cm := NewConfigManager(GetEnvLoader(""))

	testEnvKey := "TEST_ENV"
	testEnvValue := "VAL://yrl?=1"
	os.Setenv(testEnvKey, testEnvValue)

	v := cm.Get(testEnvKey)
	if v != "" {
		t.Fatal("expect empty value, got: ", v)
	}

	wg := sync.WaitGroup{}

	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := cm.LoadConfig()
			if err != nil {
				t.Error("load config error", err)
			}
		}()
	}

	wg.Wait()

	v = cm.Get(testEnvKey)
	if v != testEnvValue {
		t.Fatalf("expect %v after load env, got %v", testEnvValue, v)
	}

	newEnv := "NEW_ENV_KEY"
	newValue := "VAL2"
	os.Setenv(newEnv, newValue)

	err := cm.LoadConfig()
	if err != nil {
		t.Error("call loadconfig error", err)
	}

	v = cm.Get(newEnv)
	if v != "" {
		t.Fatalf("expect dont load new env value, wait empty value, got %v", v)
	}

	cm.PrintConfig()
}
