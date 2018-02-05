package antfarm

import (
	"context"
	"errors"
	"testing"
)

type MockProvisioner struct {
	expectCalled, startCalled, abortCalled bool // internal state

	ExpectOk  bool
	ExpectErr error

	StartErr error
}

func NewMockProvisioner(options ...func(*MockProvisioner)) *MockProvisioner {
	mp := &MockProvisioner{ExpectOk: true}
	for _, option := range options {
		option(mp)
	}
	return mp
}

func (mp *MockProvisioner) Expect() (bool, error) {
	mp.expectCalled = true
	return mp.ExpectOk, mp.ExpectErr
}

func (mp *MockProvisioner) Start(ctx context.Context) error {
	mp.startCalled = true
	return mp.StartErr
}

func (mp *MockProvisioner) Abort() { mp.abortCalled = true }

func (mp *MockProvisioner) Run(t *testing.T, expect, start, abort bool, err error) {
	if e := Provision(mp).Start(context.Background()); e != err {
		t.Errorf("should return the expected error, expected: %s, got: %s", err, e)
	}

	if mp.expectCalled != expect {
		if expect {
			t.Errorf("expecting function should have been called")
		} else {
			t.Errorf("expecting function should not have been called")
		}
	}
	if mp.startCalled != start {
		if start {
			t.Errorf("start function should have been called")
		} else {
			t.Errorf("start function should not have been called")
		}
	}
	if mp.abortCalled != abort {
		if abort {
			t.Errorf("abort function should have been called")
		} else {
			t.Errorf("abort function should not have been called")
		}
	}
}

func TestProvisioner(t *testing.T) {
	NewMockProvisioner().Run(t, true, true, false, nil)
}

func TestProvisionerExpect(t *testing.T) {
	NewMockProvisioner(func(mp *MockProvisioner) { mp.ExpectOk = false }).Run(t, true, false, false, nil)
}

func TestProvisionerExpectError(t *testing.T) {
	expectErr := errors.New("error during expect phase")
	NewMockProvisioner(func(mp *MockProvisioner) { mp.ExpectErr = expectErr }).Run(t, true, false, false, expectErr)
}

func TestProvisionerAbortTriggered(t *testing.T) {
	runErr := errors.New("error during running phase")
	NewMockProvisioner(func(mp *MockProvisioner) { mp.StartErr = runErr }).Run(t, true, true, true, runErr)
}
