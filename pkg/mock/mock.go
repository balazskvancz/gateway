package mock

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	path = "./mocks.json"

	sleepSecond   = 3
	sleepDuration = sleepSecond * time.Second
)

type MockListener interface {
	ListenForMocks(*[]MockCall)
}

type Mock struct {
	modTime time.Time

	listener MockListener
}

type MockCall struct {
	Url        string          `json:"url"`
	StatusCode int             `json:"statusCode"`
	Data       json.RawMessage `json:"data"`
}

// Creates a new instance of the mock struct.
// If it cant read the data in, returns a nil pointer.
func New(listener MockListener) *Mock {
	stat, err := os.Stat(path)

	if err != nil {
		fmt.Printf("[MOCK]: os stat err: %v\n", err)

		return nil
	}

	mocks, err := readData()

	if err != nil {
		fmt.Printf("[MOCK]: read data err: %v\n", err)

		return nil
	}

	// Notifies the listener about the newly read mocks.
	listener.ListenForMocks(mocks)

	return &Mock{
		modTime:  stat.ModTime(),
		listener: listener,
	}
}

// Watching the file for change, and reload if its needed.
func (m *Mock) WatchReload() {
	if m == nil {
		return
	}

	for {
		// In every iteration we wait for x amount of sec.
		time.Sleep(sleepDuration)

		if !m.isModified() {
			continue
		}

		// If its changed, we have to read in the data.
		mocks, err := readData()

		// If there is none error, we should inform the listener
		// about the read in data.
		if err == nil {
			m.listener.ListenForMocks(mocks)
		}
	}
}

// Reads the data from the given file.
func readData() (*[]MockCall, error) {
	b, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	mocks := []MockCall{}

	if err := json.Unmarshal(b, &mocks); err != nil {
		return nil, err
	}

	return &mocks, nil
}

// Check if the file has changed.
func (m *Mock) isModified() bool {
	if m == nil {
		return false
	}

	stat, err := os.Stat(path)

	if err != nil {
		fmt.Printf("file check err: %v\n", err)

		return false
	}

	currentModTime := stat.ModTime()

	// If it is the same, it hasnt changed.
	if m.modTime == currentModTime {
		return false
	}

	// We have to register the new time.
	m.modTime = currentModTime

	return true
}
