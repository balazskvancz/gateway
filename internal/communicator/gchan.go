package communicator

import (
	"net/http"
)

type GChan struct {
	Data   chan []byte
	Status chan int
	Header chan http.Header
}

type GChanValues struct {
	Data       []byte
	StatusCode int
	Header     http.Header
}

// Resets the GChan.
func (gchan *GChan) Reset() {
	gchan.Data = make(chan []byte)
	gchan.Status = make(chan int)
	gchan.Header = make(chan http.Header)
}

// Returns if there is more to go.
func (ch *GChan) StillToGo() bool {
	return (ch.Data != nil ||
		ch.Status != nil ||
		ch.Header != nil)
}

// Reads the chans value safely.
func (ch *GChan) GetValues() GChanValues {
	data, statusCode, header := []byte{}, http.StatusBadRequest, http.Header{}

	for ch.StillToGo() {
		select {
		case d := <-ch.Data:
			data = d
			ch.Data = nil

		case i := <-ch.Status:
			statusCode = i
			ch.Status = nil

		case h := <-ch.Header:
			header = h
			ch.Header = nil
		}
	}

	return GChanValues{
		Data:       data,
		StatusCode: statusCode,
		Header:     header,
	}
}
