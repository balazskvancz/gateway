package communicator

type concurrentChan struct {
	url chan string

	*GChan
}

type concurrentValues struct {
	url string

	GChanValues
}

type ConcurrentRequest struct {
	Client GoRequest

	Url    string
	Method string

	Data []byte
}

// The respone type of multiple asyncronous calls.
type ConcurrentResponse = map[string]GChanValues

// Creates a new instance of the concurrentChan.
func newConcurrentChan() *concurrentChan {
	gchan := &GChan{}
	gchan.Reset()

	return &concurrentChan{
		url:   make(chan string),
		GChan: gchan,
	}
}

// Sends all the requests, wait for all of them,
// and returns the data, inside a map, where the
// key is the url itsfelf.
func ConcurrentAll(requests ...ConcurrentRequest) ConcurrentResponse {
	channels := []*concurrentChan{}

	for range requests {
		ch := newConcurrentChan()

		channels = append(channels, ch)
	}

	for idx, r := range requests {
		m, u, d := r.Method, r.Url, r.Data

		ch := channels[idx]

		go r.Client.GoRequest(m, u, d, ch)
	}

	res := make(map[string]GChanValues)

	for _, ch := range channels {
		cValues := ch.GetValues()

		res[cValues.url] = cValues.GChanValues
	}

	return res
}

// Gets the values from the channel.
func (c concurrentChan) GetValues() concurrentValues {
	url := <-c.url
	gChanValues := c.GChan.GetValues()

	return concurrentValues{
		url:         url,
		GChanValues: gChanValues,
	}
}
