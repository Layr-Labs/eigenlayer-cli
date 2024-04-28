package utils

import (
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
)

// Prompter is an interface for prompting the user for input.
type Prompter interface {
	Select(prompt string, options []string) (string, error)
	InputString(prompt, defValue, help string, validator func(string) error) (string, error)
	InputInteger(prompt, defValue, help string, validator func(int64) error) (int64, error)
	Confirm(prompt string) (bool, error)
	InputHiddenString(prompt, help string, validator func(string) error) (string, error)
}

type prompter struct{}

// NewPrompter returns a new Prompter instance.
func NewPrompter() Prompter {
	return &prompter{}
}

// Select prompts the user to select one of the options provided.
func (p *prompter) Select(prompt string, options []string) (string, error) {
	selected := ""
	s := &survey.Select{
		Message: prompt,
		Options: options,
	}
	err := survey.AskOne(s, &selected)
	return selected, err
}

// InputString prompts the user to input a string. The default value is used if the user does not provide any input.
// The validator is used to validate the input. The help text is displayed to the user when they ask for help.
func (p *prompter) InputString(prompt, defValue, help string, validator func(string) error) (string, error) {
	var result string
	i := &survey.Input{
		Message: prompt,
		Default: defValue,
		Help:    help,
	}
	err := survey.AskOne(i, &result, survey.WithValidator(func(ans interface{}) error {
		if err := validator(ans.(string)); err != nil {
			return err
		}
		return nil
	}))
	return result, err
}

func (p *prompter) InputInteger(prompt, defValue, help string, validator func(int64) error) (int64, error) {
	var result int64
	i := &survey.Input{
		Message: prompt,
		Default: defValue,
		Help:    help,
	}

	err := survey.AskOne(i, &result, survey.WithValidator(func(ans interface{}) error {
		atoi, err := strconv.Atoi(ans.(string))
		if err != nil {
			return fmt.Errorf("invalid integer with err: %s", err.Error())
		}
		if err := validator(int64(atoi)); err != nil {
			return err
		}
		return nil
	}))
	return result, err
}

// Confirm prompts the user to confirm an action with a yes/no question.
func (p *prompter) Confirm(prompt string) (bool, error) {
	result := false
	c := &survey.Confirm{
		Message: prompt,
	}
	err := survey.AskOne(c, &result)
	return result, err
}

// InputHiddenString prompts the user to input a string. The input is hidden from the user.
// The validator is used to validate the input. The help text is displayed to the user when they ask for help.
// There is no default value.
func (p *prompter) InputHiddenString(prompt, help string, validator func(string) error) (string, error) {
	var result string
	i := &survey.Password{
		Message: prompt,
		Help:    help,
	}

	err := survey.AskOne(i, &result, survey.WithValidator(func(ans interface{}) error {
		if err := validator(ans.(string)); err != nil {
			return err
		}
		return nil
	}))
	return result, err
}
