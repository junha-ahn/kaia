// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kaiachain/kaia/kas (interfaces: AnchorDB)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockAnchorDB is a mock of AnchorDB interface
type MockAnchorDB struct {
	ctrl     *gomock.Controller
	recorder *MockAnchorDBMockRecorder
}

// MockAnchorDBMockRecorder is the mock recorder for MockAnchorDB
type MockAnchorDBMockRecorder struct {
	mock *MockAnchorDB
}

// NewMockAnchorDB creates a new mock instance
func NewMockAnchorDB(ctrl *gomock.Controller) *MockAnchorDB {
	mock := &MockAnchorDB{ctrl: ctrl}
	mock.recorder = &MockAnchorDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAnchorDB) EXPECT() *MockAnchorDBMockRecorder {
	return m.recorder
}

// ReadAnchoredBlockNumber mocks base method
func (m *MockAnchorDB) ReadAnchoredBlockNumber() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadAnchoredBlockNumber")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// ReadAnchoredBlockNumber indicates an expected call of ReadAnchoredBlockNumber
func (mr *MockAnchorDBMockRecorder) ReadAnchoredBlockNumber() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadAnchoredBlockNumber", reflect.TypeOf((*MockAnchorDB)(nil).ReadAnchoredBlockNumber))
}

// WriteAnchoredBlockNumber mocks base method
func (m *MockAnchorDB) WriteAnchoredBlockNumber(arg0 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WriteAnchoredBlockNumber", arg0)
}

// WriteAnchoredBlockNumber indicates an expected call of WriteAnchoredBlockNumber
func (mr *MockAnchorDBMockRecorder) WriteAnchoredBlockNumber(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteAnchoredBlockNumber", reflect.TypeOf((*MockAnchorDB)(nil).WriteAnchoredBlockNumber), arg0)
}
