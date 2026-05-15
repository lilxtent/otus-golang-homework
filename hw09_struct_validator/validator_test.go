package hw09structvalidator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			// Place your code here.
		},
		// ...
		// Place your code here.
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			// Place your code here.
			_ = tt
		})
	}
}

func TestValidCases(t *testing.T) {
	validCases := []struct {
		value    any
		tagValue string
	}{
		{value: 15, tagValue: `min:10`},
		{value: 5, tagValue: `max:10`},
		{value: 12, tagValue: `in:10,15`},
		{value: 12, tagValue: `min:10|max:15`},
		{value: "123", tagValue: `len:3`},
		{value: "123", tagValue: `regexp:\d+`},
		{value: "foo", tagValue: `in:foo,bar`},
		{value: "foo", tagValue: `len:3|in:foo,bar`},
	}

	for _, tc := range validCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := ValidateFieldValue(tc.value, tc.tagValue)

			require.Empty(t, errors)
		})
	}
}

func TestInvalidCases(t *testing.T) {
	invalidCases := []struct {
		value    any
		tagValue string
	}{
		{value: 5, tagValue: `min:10`},
		{value: 15, tagValue: `max:10`},
		{value: 5, tagValue: `in:10,15`},
		{value: 20, tagValue: `in:10,15`},
		{value: 5, tagValue: `min:10|max:15`},
		{value: "1234", tagValue: `len:3`},
		{value: "abc", tagValue: `regexp:\d+`},
		{value: "unexpected", tagValue: `in:foo,bar`},
		{value: "sixseven", tagValue: `len:3|in:foo,bar`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := ValidateFieldValue(tc.value, tc.tagValue)

			require.NotEmpty(t, errors)
		})
	}
}

func TestValidSliceCases(t *testing.T) {
	validCases := []struct {
		values   []any
		tagValue string
	}{
		{values: []any{15, 100, 150}, tagValue: `min:10`},
		{values: []any{0, 5, 9}, tagValue: `max:10`},
		{values: []any{12, 13, 14}, tagValue: `in:10,15`},
		{values: []any{12, 13, 14}, tagValue: `min:10|max:15`},
		{values: []any{"123", "345"}, tagValue: `len:3`},
		{values: []any{"123", "345"}, tagValue: `regexp:\d+`},
		{values: []any{"foo", "bar"}, tagValue: `in:foo,bar`},
		{values: []any{"123", "345"}, tagValue: `len:3|in:123,345`},
	}

	for _, tc := range validCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := ValidateFieldValueSlice(tc.values, tc.tagValue)

			require.Empty(t, errors)
		})
	}
}

func TestInvalidSliceCases(t *testing.T) {
	invalidCases := []struct {
		values   []any
		tagValue string
	}{
		{values: []any{5, 9}, tagValue: `min:10`},
		{values: []any{15, 100}, tagValue: `max:10`},
		{values: []any{5, 20}, tagValue: `in:10,15`},
		{values: []any{5, 20}, tagValue: `min:10|max:15`},
		{values: []any{"123", "3456"}, tagValue: `len:3`},
		{values: []any{"abc", "345"}, tagValue: `regexp:\d+`},
		{values: []any{"foo", "notfoo"}, tagValue: `in:foo,bar`},
		{values: []any{"123", "3456"}, tagValue: `len:3|in:123,3456`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := ValidateFieldValueSlice(tc.values, tc.tagValue)

			require.NotEmpty(t, errors)
		})
	}
}
