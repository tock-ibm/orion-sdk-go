// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// SignatureVerifier is an autogenerated mock type for the SignatureVerifier type
type SignatureVerifier struct {
	mock.Mock
}

// Verify provides a mock function with given fields: entityID, payload, signature
func (_m *SignatureVerifier) Verify(entityID string, payload []byte, signature []byte) error {
	ret := _m.Called(entityID, payload, signature)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte, []byte) error); ok {
		r0 = rf(entityID, payload, signature)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
