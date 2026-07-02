package queue

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	message, err := MarshalJSON(struct {
		Title string `json:"title"`
	}{
		Title: "standup",
	})
	require.NoError(t, err)
	require.Equal(t, ContentTypeJSON, message.ContentType)
	require.JSONEq(t, `{"title":"standup"}`, string(message.Body))
}

func TestMarshalJSONReturnsError(t *testing.T) {
	t.Parallel()

	_, err := MarshalJSON(func() {})
	require.Error(t, err)
}

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var payload struct {
		Title string `json:"title"`
	}

	err := UnmarshalJSON(Message{Body: []byte(`{"title":"standup"}`)}, &payload)
	require.NoError(t, err)
	require.Equal(t, "standup", payload.Title)
}

func TestUnmarshalJSONReturnsError(t *testing.T) {
	t.Parallel()

	var payload struct {
		Title string `json:"title"`
	}

	err := UnmarshalJSON(Message{Body: []byte(`{`)}, &payload)
	require.Error(t, err)
}

func TestUnmarshalJSONReturnsTypeError(t *testing.T) {
	t.Parallel()

	var payload struct {
		Count int `json:"count"`
	}

	err := UnmarshalJSON(Message{Body: []byte(`{"count":"many"}`)}, &payload)
	var typeErr *json.UnmarshalTypeError
	require.True(t, errors.As(err, &typeErr))
}
