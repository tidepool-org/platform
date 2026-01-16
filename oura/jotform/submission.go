package jotform

import (
	"encoding/json"
	"fmt"
)

type SubmissionResponse struct {
	ResponseCode int     `json:"responseCode"`
	Content      Content `json:"content"`
}

type Content struct {
	ID        string  `json:"id"`
	FormID    string  `json:"form_id"`
	Answers   Answers `json:"answers"`
	CreatedAt string  `json:"created_at"`
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

type ControlDateTime struct {
	BaseAnswer
	AnswerField  DateTimeAnswerField `json:"answer"`
	PrettyFormat string              `json:"prettyFormat"`
}

func (c ControlDateTime) Name() string {
	return c.NameField
}

func (c ControlDateTime) Answer() string {
	return c.AnswerField.DateTime
}

type DateTimeAnswerField struct {
	Year     string `json:"year"`
	Month    string `json:"month"`
	Day      string `json:"day"`
	DateTime string `json:"datetime"`
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
		case "control_datetime":
			var cd ControlDateTime
			if err := json.Unmarshal(rawMsg, &cd); err != nil {
				return fmt.Errorf("failed to unmarshal ControlDateTime for key %s: %w", key, err)
			}
			answer = cd
		default:
			continue
		}

		(*a)[answer.Name()] = answer
	}

	return nil
}
