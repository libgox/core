package dynamic

import (
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/libgox/core/config/dynamic/types"
)

type ConfigManager[T any] struct {
	options       options
	source        types.Source
	parseFunc     func([]byte) (T, error) // Parsing function to convert []byte to T
	currentConfig T
	listeners     []types.Listener[T]
	mu            sync.RWMutex
	stopChan      chan struct{}
}

// NewConfigManager creates a new ConfigManager instance with a custom parsing function
func NewConfigManager[T any](source types.Source, parseFunc func([]byte) (T, error), options ...Option) *ConfigManager[T] {
	cm := &ConfigManager[T]{
		source:    source,
		parseFunc: parseFunc,
		stopChan:  make(chan struct{}),
	}

	//Default poll interval
	options = append(options, WithPollInterval(time.Second))

	for _, f := range options {
		f(&cm.options)
	}

	return cm
}

func (cm *ConfigManager[T]) Start() error {
	if cm.source.Type() == types.Polling {
		go cm.startPolling()
	} else {
		cm.source.SetUpdateCallback(cm.handleUpdate)
		return cm.source.Start()
	}

	return nil
}

func (cm *ConfigManager[T]) Stop() {
	_ = cm.source.Stop()
	close(cm.stopChan)
}

func (cm *ConfigManager[T]) startPolling() {
	ticker := time.NewTicker(cm.options.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cm.pollConfig()
		case <-cm.stopChan:
			return
		}
	}
}

// pollConfig performs a single poll
func (cm *ConfigManager[T]) pollConfig() {
	configBytes, err := cm.source.Read()
	if err != nil {
		log.Printf("Error reading config: %v", err)
		return
	}
	cm.handleUpdate(configBytes)
}

// handleUpdate handles configuration updates using the provided parseFunc
func (cm *ConfigManager[T]) handleUpdate(configBytes []byte) {
	newConfig, err := cm.parseFunc(configBytes)
	if err != nil {
		log.Printf("Failed to parse config: %v", err)
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !reflect.DeepEqual(cm.currentConfig, newConfig) {
		cm.currentConfig = newConfig
		for _, listener := range cm.listeners {
			go listener.Update(newConfig)
		}
	}
}

func (cm *ConfigManager[T]) RegisterListener(listener types.Listener[T]) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.listeners = append(cm.listeners, listener)
}
