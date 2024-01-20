package types

import (
	"fmt"
)

type Question struct {
	ID               uint           `json:"id"`
	ModuleID         *uint          `json:"module_id,omitempty"`
	QuestionnaireID  *uint          `json:"questionnaire_id,omitempty"`
	TextRoot         string         `json:"text_root"`
	Difficulty       int            `json:"difficulty"`
	TypeQuestion     string         `json:"type_question" validate:"required,oneof=true_false multi_choice_text multi_choice_abc complete_word order_word"`
	QuestionAnswerID uint           `json:"question_answer_id,omitempty"`
	QuestionAnswer   QuestionAnswer `json:"question_answer,omitempty"`
	CorrectAnswerID  uint           `json:"correct_answer_id,omitempty"`
	CorrectAnswer    Answer         `json:"correct_answer,omitempty"`
}

func (q *Question) Validate() error {
	if q.TextRoot == "" {
		return fmt.Errorf("the text root cannot be empty")
	}

	if q.TypeQuestion == "" {
		return fmt.Errorf("the type question cannot be empty")
	}

	if q.TypeQuestion == "multi_choice_text" || q.TypeQuestion == "multi_choice_abc" {
		if len(q.QuestionAnswer.TextOptions) == 0 {
			return fmt.Errorf("the text options cannot be empty")
		}

		if q.QuestionAnswer.SelectMode == "" {
			return fmt.Errorf("the select mode cannot be empty")
		}

		if q.QuestionAnswer.SelectMode != "single" && q.QuestionAnswer.SelectMode != "multiple" {
			return fmt.Errorf("the select mode must be single or multiple")
		}

		for _, option := range q.QuestionAnswer.TextOptions {
			if option == "" {
				return fmt.Errorf("the text options cannot be empty")
			}
		}

		// validar que la respuesta este dentro de las opciones.
		ok := false
		for _, option := range q.QuestionAnswer.TextOptions {
			if option == q.CorrectAnswer.TextOpcions[0] {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("the correct answer must be one of the options")
		}
	}

	if q.TypeQuestion == "complete_word" {
		if q.QuestionAnswer.TextToComplete == "" {
			return fmt.Errorf("the text to complete cannot be empty")
		}

		if len(q.CorrectAnswer.TextToComplete) == 0 {
			return fmt.Errorf("the correct answer cannot be empty")
		}
	}

	if q.TypeQuestion == "order_word" {
		if len(q.QuestionAnswer.TextOptions) == 0 {
			return fmt.Errorf("the text to complete cannot be empty")
		}

		if len(q.CorrectAnswer.TextOpcions) == 0 {
			return fmt.Errorf("the correct answer cannot be empty")
		}
	}

	return nil
}

type QuestionAnswer struct {
	ID             uint     `json:"id"`
	SelectMode     string   `json:"select_mode" validate:"required,oneof=single multiple"`
	TextOptions    []string `json:"text_options"`
	TextToComplete string   `json:"text_to_complete"`
	Hind           string   `json:"hind"`
}

type Answer struct {
	ID             uint     `json:"id"`
	TrueOrFalse    bool     `json:"true_or_false"`
	TextOpcions    []string `json:"text_opcions"`
	TextToComplete []string `json:"text_to_complete"`
}
