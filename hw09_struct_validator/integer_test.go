package hw09structvalidator

import (
	"testing"

	"github.com/fixme_my_friend/hw09_struct_validator/validators"
	"github.com/stretchr/testify/require"
)

func TestValidCases(t *testing.T){
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

func TestInvalidCases(t *testing.T){
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