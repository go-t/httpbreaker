package httpbreaker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
)

type testTransport struct {
	res *http.Response
	err error
}

func (tr testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return tr.res, tr.err
}

func TestSetup(t *testing.T) {
	openTimeout := 100 * time.Millisecond
	maxFailures := 2

	cfg := gobreaker.Settings{
		Name:        "httpbreaker",
		MaxRequests: 2,
		Timeout:     openTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > uint32(maxFailures)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			t.Logf("state %s -> %s", from, to)
		},
	}

	res := httptest.NewRecorder()
	res.WriteString("OK")

	tr := &testTransport{
		res: res.Result(),
		err: fmt.Errorf("http error"),
	}

	client := Wrap(&http.Client{Transport: tr}, cfg)

	_, err := client.Get("some_url")
	assert.Error(t, err)
	assert.Equal(t, client.State(), gobreaker.StateClosed)

	_, err = client.Get("some_url")
	assert.Error(t, err)
	assert.Equal(t, client.State(), gobreaker.StateClosed)

	_, err = client.Get("some_url")
	assert.Error(t, err)
	assert.Equal(t, client.State(), gobreaker.StateOpen)

	time.Sleep(150 * time.Millisecond)
	tr.err = nil

	_, err = client.Get("some_url")
	assert.NoError(t, err)
	assert.Equal(t, client.State(), gobreaker.StateHalfOpen)

	_, err = client.Get("some_url")
	assert.NoError(t, err)
	assert.Equal(t, client.State(), gobreaker.StateClosed)

}
