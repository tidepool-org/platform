package jotform

import (
	"encoding/json"
	"fmt"
)

type SubmissionResponse struct {
	Content Content `json:"content"`
}

type Content struct {
	ID      string  `json:"id"`
	Answers Answers `json:"answers"`
}

type Answer interface {
	Name() string
	Answer() string
}

type BaseAnswer struct {
	NameField string `json:"name"`
	Order     string `json:"order"`
	Text      string `json:"text"`
	Type      string `json:"type"`
	Sublabels string `json:"sublabels,omitempty"`
}

type ControlTextbox struct {
	BaseAnswer
	AnswerField string `json:"answer"`
}

func (c ControlTextbox) Name() string {
	return c.NameField
}

func (c ControlTextbox) Answer() string {
	return c.AnswerField
}

type ControlFullname struct {
	BaseAnswer
	AnswerField  map[string]string `json:"answer"`
	PrettyFormat string            `json:"prettyFormat"`
}

func (c ControlFullname) Name() string {
	return c.NameField
}

func (c ControlFullname) Answer() string {
	return c.PrettyFormat
}

type Answers map[string]Answer

func (a Answers) GetAnswerTextByName(name string) string {
	answer, ok := a[name]
	if !ok {
		return ""
	}
	return answer.Answer()
}

// UnmarshalJSON implements custom unmarshaling for AnswersMap
func (a *Answers) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*a = make(map[string]Answer)

	for key, rawMsg := range raw {
		var typeInfo struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(rawMsg, &typeInfo); err != nil {
			return fmt.Errorf("failed to unmarshal type for key %s: %w", key, err)
		}

		var answer Answer
		switch typeInfo.Type {
		case "control_textbox":
			var ct ControlTextbox
			if err := json.Unmarshal(rawMsg, &ct); err != nil {
				return fmt.Errorf("failed to unmarshal ControlTextbox for key %s: %w", key, err)
			}
			answer = ct
		case "control_fullname":
			var cf ControlFullname
			if err := json.Unmarshal(rawMsg, &cf); err != nil {
				return fmt.Errorf("failed to unmarshal ControlFullname for key %s: %w", key, err)
			}
			answer = cf
		default:
			return fmt.Errorf("unknown answer type: %s for key %s", typeInfo.Type, key)
		}

		(*a)[answer.Name()] = answer
	}

	return nil
}
