package gcontext

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	stateStarted = "STARTED"
	stateEnded   = "ENDED"

	filePath = "gateway.log"
)

type logger struct {
	mu sync.Mutex
}

// Creates a new logger instance.
func NewLogger() *logger {
	if !isFileExists() {
		if err := createLogFile(); err != nil {
			fmt.Printf("[CONTEXT]: cant create file: %v\n", err)
			return nil
		}
	}

	return &logger{
		mu: sync.Mutex{},
	}
}

// Writes to log file.
func (l *logger) writeToLog(id uint64, state, action string) {
	if l == nil {
		return
	}

	// At this point, if the file doesnt exist, we shouldnt do anything.
	if !isFileExists() {
		fmt.Printf("[CONTEXT]: unable to write log file. File: %s does not exist.\n", filePath)

		return
	}

	// Acquire the lock, so it wont be a problem,
	// if there is more than one thread to write it.
	l.mu.Lock()

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		fmt.Printf("[CONTEXT]: log error %v\n", err)

		return
	}

	dt := time.Now()

	// Creating the current events log.
	strLog := fmt.Sprintf("%s\t%d\t%s\t%s\n", dt.Format("2006:01:02 15:04:05.999999999"), id, state, action)

	file.Write([]byte(strLog))

	// Cleaning up the allocated resources
	file.Close()
	l.mu.Unlock()
}

// Checks whether the log file exists.
func isFileExists() bool {
	_, err := os.Stat(filePath)

	return err == nil
}

// Tries to crete the log file.
func createLogFile() error {
	f, err := os.Create(filePath)

	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}
