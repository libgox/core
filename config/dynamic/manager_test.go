package dynamic

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/libgox/core/config/dynamic/parse"
	"github.com/libgox/core/config/dynamic/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSource struct {
	mock.Mock
	updateCallback func([]byte)
}

func (m *MockSource) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSource) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSource) Type() types.SourceType {
	args := m.Called()
	return args.Get(0).(types.SourceType)
}

func (m *MockSource) Read() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSource) SetUpdateCallback(callback func([]byte)) {
	m.Called(callback)
	m.updateCallback = callback
}

type MockListener[T any] struct {
	mock.Mock
	wg sync.WaitGroup
}

func (m *MockListener[T]) Update(config T) {
	m.Called(config)
	m.wg.Done()
}

// Wait waits for all expected calls to complete
func (m *MockListener[T]) Wait() {
	m.wg.Wait()
}

// AddWaitCount adds count to wait group
func (m *MockListener[T]) AddWaitCount(count int) {
	m.wg.Add(count)
}

type ConfigData struct {
	Level string `json:"level"`
}

func TestConfigManager(t *testing.T) {
	mockSource := new(MockSource)
	pollInterval := time.Second
	manager := NewConfigManager(
		mockSource,
		parse.JSONParseFunc[ConfigData],
		WithPollInterval(pollInterval),
	)

	assert.NotNil(t, manager)
	assert.Equal(t, mockSource, manager.source)
	assert.Equal(t, pollInterval, manager.options.pollInterval)
	assert.Equal(t,
		reflect.ValueOf(parse.JSONParseFunc[ConfigData]).Pointer(),
		reflect.ValueOf(manager.parseFunc).Pointer())
	assert.NotNil(t, manager.stopChan)
}

func TestConfigManager_Start_Polling(t *testing.T) {
	jsonData := []byte(`{"level":"debug"}`)
	mockSource := new(MockSource)
	mockSource.On("Type").Return(types.Polling)
	mockSource.On("Stop").Return(nil)
	mockSource.On("Read").Return(jsonData, nil)

	mockListener := new(MockListener[ConfigData])
	mockListener.On("Update", ConfigData{Level: "debug"}).Times(1)
	mockListener.AddWaitCount(1)

	manager := NewConfigManager(
		mockSource,
		parse.JSONParseFunc[ConfigData],
		WithPollInterval(time.Millisecond*100),
	)
	manager.RegisterListener(mockListener)

	err := manager.Start()
	assert.NoError(t, err)
	mockListener.Wait()

	manager.Stop()
	mockSource.AssertExpectations(t)
	mockListener.AssertExpectations(t)
}

func TestConfigManager_Start_Callback(t *testing.T) {
	jsonData := []byte(`{"level":"debug"}`)
	mockSource := new(MockSource)
	mockSource.On("Type").Return(types.Dynamic)
	mockSource.On("Stop").Return(nil)
	mockSource.On("Start").Return(nil)
	mockSource.On("SetUpdateCallback", mock.AnythingOfType("func([]uint8)"))

	mockListener := new(MockListener[ConfigData])
	mockListener.On("Update", ConfigData{Level: "debug"}).Times(1)
	mockListener.AddWaitCount(1)

	manager := NewConfigManager(mockSource, parse.JSONParseFunc[ConfigData])
	manager.RegisterListener(mockListener)
	err := manager.Start()
	assert.NoError(t, err)

	mockSource.updateCallback(jsonData)
	mockListener.Wait()

	manager.Stop()
	mockSource.AssertExpectations(t)
	mockListener.AssertExpectations(t)
}
