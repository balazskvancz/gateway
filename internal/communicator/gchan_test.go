package communicator

import (
	"net/http"
	"testing"
	"time"
)

func TestGChanMerge(t *testing.T) {
	// Basic test, for merging 3 channels together.
	t.Run("testing the merge of the gchan", func(t *testing.T) {
		gch := new(GChan)
		gch.Reset()

		go func(ch *GChan) {
			ch.Status <- http.StatusOK
			time.Sleep(1 * time.Second)
			ch.Data <- []byte{}
			ch.Header <- nil
		}(gch)

		values := gch.GetValues()

		if len(values.Data) != 0 {
			t.Errorf("expected length: 0; got: %d\n", len(values.Data))
		}

		if values.StatusCode != http.StatusOK {
			t.Errorf("expected statusCode: %d; got: %d\n", http.StatusOK, values.StatusCode)
		}

		if values.Header != nil {
			t.Errorf("expected header to be nil\n")
		}
	})

}
