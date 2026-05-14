package hw09structvalidator

import (
	"testing"

	"github.com/fixme_my_friend/hw09_struct_validator/validators"
	"github.com/stretchr/testify/require"
)

func TestValidIntCases(t *testing.T){
	validCases := []struct{
	value int64 
	tagValue string
	}{
		{value: 15, tagValue: `min:10`},
		{value: 5, tagValue: `max:10`},
		{value: 12, tagValue: `in:10,15`},
		{value: 12, tagValue: `min:10|max:15`},
	}

	for _, tc := range validCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := validators.ValidateInt(tc.value, tc.tagValue)

			require.Empty(t, errors)
		})
	} 
}

func TestInvalidIntCases(t *testing.T){
	invalidCases := []struct{
	value int64 
	tagValue string
	}{
		{value: 5, tagValue: `min:10`},
		{value: 15, tagValue: `max:10`},
		{value: 5, tagValue: `in:10,15`},
		{value: 20, tagValue: `in:10,15`},
		{value: 5, tagValue: `min:10|max:15`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := validators.ValidateInt(tc.value, tc.tagValue)

			require.NotEmpty(t, errors)
		})
	} 
}

func TestValidIntSliceCases(t *testing.T){
	validCases := []struct{
	values []int64 
	tagValue string
	}{
		{values: []int64{15, 100, 150}, tagValue: `min:10`},
		{values: []int64{0, 5, 9}, tagValue: `max:10`},
		{values: []int64{12, 13, 14}, tagValue: `in:10,15`},
		{values: []int64{12, 13 ,14}, tagValue: `min:10|max:15`},
	}

	for _, tc := range validCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := validators.ValidateIntSlice(tc.values, tc.tagValue)

			require.Empty(t, errors)
		})
	} 
}

func TestInvalidIntSliceCases(t *testing.T){
	invalidCases := []struct{
	values []int64 
	tagValue string
	}{
		{values: []int64{5, 9}, tagValue: `min:10`},
		{values: []int64{15, 100}, tagValue: `max:10`},
		{values: []int64{5, 20}, tagValue: `in:10,15`},
		{values: []int64{5, 20}, tagValue: `min:10|max:15`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := validators.ValidateIntSlice(tc.values, tc.tagValue)

			require.NotEmpty(t, errors)
		})
	} 
}