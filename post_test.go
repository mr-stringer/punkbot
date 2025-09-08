package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockTokenManager struct{}

//	getToken(cnf *Config, url string) (*DIDResponse, error)
//	getRefresh(current **DIDResponse, url string) error

func (m *mockTokenManager) getToken(cnf *Config, url string) (*DIDResponse, error) {
	var ret DIDResponse
	err := json.Unmarshal([]byte(goodJson), &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (m *mockTokenManager) getRefresh(current **DIDResponse, url string) error {
	return nil
}

type mockTokenManagerBadGet struct{}

func (m *mockTokenManagerBadGet) getToken(cnf *Config, url string) (*DIDResponse, error) {
	return nil, fmt.Errorf("couldn't get token")
}

func (m *mockTokenManagerBadGet) getRefresh(current **DIDResponse, url string) error {
	return nil
}

type mockTokenManagerBadRefresh struct{}

func (m *mockTokenManagerBadRefresh) getToken(cnf *Config, url string) (*DIDResponse, error) {
	var ret DIDResponse
	err := json.Unmarshal([]byte(goodJson), &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (m *mockTokenManagerBadRefresh) getRefresh(current **DIDResponse, url string) error {
	return fmt.Errorf("cannot refresh token")
}

func Test_sessionServer(t *testing.T) {

	mtm := mockTokenManager{}
	cfg := Config{Identifier: "did:plc:acb", password: "blah-blah"}
	cp := ChanPkg{
		ByteSlice:      make(chan []byte, ByteSliceBufferSize),
		ReqDidResp:     make(chan bool),
		Session:        make(chan DIDResponse),
		JetStreamError: make(chan bool),
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tm  TokenManagerInt
		wg  *sync.WaitGroup
		cnf *Config
		cp  ChanPkg
		tr  time.Duration
	}{
		{"Good01", &mtm, &sync.WaitGroup{}, &cfg, cp, time.Millisecond * 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			tt.wg.Add(1)
			go sessionServer(tt.tm, ctx, tt.wg, tt.cnf, tt.cp, tt.tr)
			time.Sleep(time.Second * 1)
			cancel()
			tt.wg.Wait()
		})
	}
}

func Test_sessionServer_BadGet(t *testing.T) {
	/* override exit function */
	bttm := mockTokenManagerBadGet{}
	cnf := Config{Identifier: "did:plc:acb", password: "blah-blah"}
	cp := ChanPkg{
		ByteSlice:      make(chan []byte, ByteSliceBufferSize),
		ReqDidResp:     make(chan bool),
		Session:        make(chan DIDResponse),
		JetStreamError: make(chan bool),
		Exit:           make(chan int, 1),
		/* the use of a buffered channel for exit allow the test to be less.  */
		/* complex and helps avoid a race condition in the final comparison   */
		/* but a buffered channel is not need in normal operation             */
	}
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	//run as go routine
	go func() {
		go sessionServer(&bttm, ctx, &wg, &cnf, cp, time.Millisecond*100)
	}()

	/* Look for something out of the Exit channel*/
	var i = 0

	select {
	case i = <-cp.Exit:
	case <-time.After(time.Millisecond * 250):
		t.Errorf("expected exit code from channel but timed out")
	}

	wg.Wait()
	if i != ExitGetToken {
		t.Errorf("expected exit code, %d, but got %d", ExitGetToken, i)
	}

}

func Test_sessionServer_BadRefresh(t *testing.T) {
	/* override timeouts for testing*/
	TokenRefreshTimeoutMs = 5

	brtm := mockTokenManagerBadRefresh{}
	cnf := Config{Identifier: "did:plc:acb", password: "blah-blah"}
	cp := ChanPkg{
		ByteSlice:      make(chan []byte, ByteSliceBufferSize),
		ReqDidResp:     make(chan bool),
		Session:        make(chan DIDResponse),
		JetStreamError: make(chan bool),
		Exit:           make(chan int, 1),
		/* the use of a buffered channel for exit allow the test to be less.  */
		/* complex and helps avoid a race condition in the final comparison   */
		/* but a buffered channel is not need in normal operation             */
	}
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	//run as go routine
	go func() {
		go sessionServer(&brtm, ctx, &wg, &cnf, cp, time.Millisecond*10)
	}()

	var i = 0

	select {
	case i = <-cp.Exit:
	case <-time.After(time.Millisecond * 250):
		t.Errorf("expected exit code from channel but timed out")
	}

	wg.Wait()
	if i != ExitRefreshToken {
		t.Errorf("expected exit code, %d, but got %d", ExitRefreshToken, i)
	}

}
