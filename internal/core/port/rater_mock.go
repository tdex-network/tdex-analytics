// Code generated by mockery v2.10.0. DO NOT EDIT.

package port

import (
	context "context"

	decimal "github.com/shopspring/decimal"
	mock "github.com/stretchr/testify/mock"
)

// MockRateService is an autogenerated mock type for the RateService type
type MockRateService struct {
	mock.Mock
}

// ConvertCurrency provides a mock function with given fields: ctx, source, target
func (_m *MockRateService) ConvertCurrency(ctx context.Context, source string, target string) (decimal.Decimal, error) {
	ret := _m.Called(ctx, source, target)

	var r0 decimal.Decimal
	if rf, ok := ret.Get(0).(func(context.Context, string, string) decimal.Decimal); ok {
		r0 = rf(ctx, source, target)
	} else {
		r0 = ret.Get(0).(decimal.Decimal)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, source, target)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAssetCurrency provides a mock function with given fields: assetId
func (_m *MockRateService) GetAssetCurrency(assetId string) (string, error) {
	ret := _m.Called(assetId)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(assetId)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(assetId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsFiatSymbolSupported provides a mock function with given fields: symbol
func (_m *MockRateService) IsFiatSymbolSupported(symbol string) (bool, error) {
	ret := _m.Called(symbol)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(symbol)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(symbol)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
