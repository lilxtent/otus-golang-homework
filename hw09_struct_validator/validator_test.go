package hw09structvalidator

import (
	"encoding/json"
	stdErrors "errors"
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

	Meta struct {
		CreatedBy string `validate:"len:5"`
		Score     int    `validate:"min:1"`
	}

	UserWithMeta struct {
		Meta Meta `validate:"nested"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name              string
		in                any
		expectedErrFields []string
	}{
		{
			name: "valid user",
			in: User{
				ID:     "550e8400-e29b-41d4-a716-446655440000",
				Name:   "Ivan",
				Age:    30,
				Email:  "ivan@example.com",
				Role:   UserRole("admin"),
				Phones: []string{"79991234567"},
			},
		},
		{
			name: "invalid user",
			in: User{
				ID:     "not-uuid",
				Name:   "Ivan",
				Age:    17,
				Email:  "ivan.example.com",
				Role:   UserRole("guest"),
				Phones: []string{"79991234567", "123"},
			},
			expectedErrFields: []string{"ID", "Age", "Email", "Role", "Phones"},
		},
		{
			name: "valid app",
			in:   App{Version: "1.0.0"},
		},
		{
			name:              "invalid app",
			in:                App{Version: "1.0"},
			expectedErrFields: []string{"Version"},
		},
		{
			name: "token without validation tags",
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
		},
		{
			name: "valid response",
			in:   Response{Code: 500, Body: "server error"},
		},
		{
			name:              "invalid response",
			in:                Response{Code: 201, Body: "created"},
			expectedErrFields: []string{"Code"},
		},
		{
			name: "valid nested struct",
			in: UserWithMeta{
				Meta: Meta{
					CreatedBy: "admin",
					Score:     1,
				},
			},
		},
		{
			name: "invalid nested struct",
			in: UserWithMeta{
				Meta: Meta{
					CreatedBy: "root",
					Score:     0,
				},
			},
			expectedErrFields: []string{"CreatedBy", "Score"},
		},
		{
			name: "nil",
			in:   nil,
		},
		{
			name: "non-struct",
			in:   "not a struct",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d %s", i, tt.name), func(t *testing.T) {
			tt := tt
			t.Parallel()

			errors := Validate(tt.in)

			require.Len(t, errors, len(tt.expectedErrFields))
			for i, expectedField := range tt.expectedErrFields {
				require.IsType(t, ValidationError{}, errors[i])
				require.Equal(t, expectedField, errors[i].Field)
				require.Error(t, errors[i].Err)
			}
		})
	}
}

func TestValidateInvalidTagDeclaration(t *testing.T) {
	t.Parallel()

	in := struct {
		Age    int      `validate:"min:old"`
		Email  string   `validate:"regexp:["`
		Phones []string `validate:"len:eleven"`
	}{
		Age:    30,
		Email:  "ivan@example.com",
		Phones: []string{"79991234567"},
	}

	errors := Validate(in)

	require.Len(t, errors, 3)
	for _, err := range errors {
		require.IsType(t, ValidationError{}, err)

		var tagDeclarationError *TagDeclarationError
		require.True(t, stdErrors.As(err.Err, &tagDeclarationError))
	}
}

func requireInvalidValueErrors(t *testing.T, errors []error) {
	t.Helper()
	for _, err := range errors {
		require.IsType(t, &InvalidValueError{}, err)
	}
}

func requireTagDeclarationErrors(t *testing.T, errors []error) {
	t.Helper()
	for _, err := range errors {
		var tagDeclarationError *TagDeclarationError
		require.True(t, stdErrors.As(err, &tagDeclarationError))
	}
}

func TestValidCases(t *testing.T) {
	validCases := []struct {
		value    any
		tagValue string
	}{
		{value: 15, tagValue: `min:10`},
		{value: 5, tagValue: `max:10`},
		{value: 10, tagValue: `in:10,15`},
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
		{value: 12, tagValue: `in:10,15`},
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
			requireInvalidValueErrors(t, errors)
		})
	}
}

func TestInvalidTagDeclarationCases(t *testing.T) {
	invalidCases := []struct {
		value    any
		tagValue string
	}{
		{value: 15, tagValue: `min:ten`},
		{value: 15, tagValue: `max:ten`},
		{value: 15, tagValue: `in:10,twenty`},
		{value: 15, tagValue: `between:10,20`},
		{value: "foo", tagValue: `len:three`},
		{value: "foo", tagValue: `regexp:[`},
		{value: "foo", tagValue: `contains:bar`},
		{value: "foo", tagValue: `len`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := ValidateFieldValue(tc.value, tc.tagValue)

			require.NotEmpty(t, errors)
			requireTagDeclarationErrors(t, errors)
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
		{values: []any{10, 15}, tagValue: `in:10,15`},
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

func TestInvalidSliceTagDeclarationCases(t *testing.T) {
	invalidCases := []struct {
		values   []any
		tagValue string
	}{
		{values: []any{15, 20}, tagValue: `min:ten`},
		{values: []any{15, 20}, tagValue: `in:10,twenty`},
		{values: []any{"foo", "bar"}, tagValue: `len:three`},
		{values: []any{"foo", "bar"}, tagValue: `regexp:[`},
		{values: []any{"foo", "bar"}, tagValue: `contains:foo`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			errors := ValidateFieldValueSlice(tc.values, tc.tagValue)

			require.NotEmpty(t, errors)
			requireTagDeclarationErrors(t, errors)
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
			requireInvalidValueErrors(t, errors)
		})
	}
}
