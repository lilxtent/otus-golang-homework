package queue

import "encoding/json"

const ContentTypeJSON = "application/json"

func MarshalJSON(value any) (Message, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Body:        body,
		ContentType: ContentTypeJSON,
	}, nil
}

func UnmarshalJSON(message Message, value any) error {
	return json.Unmarshal(message.Body, value)
}
