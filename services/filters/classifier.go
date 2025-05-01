package filters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mudler/LocalAGI/core/state"
	"github.com/mudler/LocalAGI/core/types"
	"github.com/mudler/LocalAGI/pkg/config"
	"github.com/mudler/LocalAGI/pkg/llm"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const FilterClassifier = "regex"

type ClassifierFilter struct {
	name         string
	client       *openai.Client
	model        string
	assertion    string
	allowOnMatch bool
	isTrigger    bool
}

type ClassifierFilterConfig struct {
	Name         string `json:"name"`
	Model        string `json:"model,omitempty"`
	Assertion    string `json:"assertion"`
	AllowOnMatch bool   `json:"allow_on_match"`
	IsTrigger    bool   `json:"is_trigger"`
}

func NewClassifierFilter(configJSON string, a *state.AgentConfig) (*ClassifierFilter, error) {
	var cfg ClassifierFilterConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, err
	}
	var model string
	if cfg.Model != "" {
		model = cfg.Model
	} else {
		model = a.Model
	}
	if cfg.Name == "" {
	  return nil, fmt.Errorf("Classifier with no name")
	}
	if cfg.Assertion == "" {
	  return nil, fmt.Errorf("%s classifier has not assertion", cfg.Name)
	}
	client := llm.NewClient(a.APIKey, a.APIURL, "1m")

  return &ClassifierFilter{
		name:         cfg.Name,
		model:        model,
		assertion:    cfg.Assertion,
		client:       client,
		allowOnMatch: cfg.AllowOnMatch,
		isTrigger:    cfg.IsTrigger,
	}, nil
}

const fmtT = `
# Assertion about message
%s

# Message
%s
`

func (f *ClassifierFilter) Name() string { return f.name }
func (f *ClassifierFilter) Apply(job *types.Job) (bool, error) {
	input := extractInputFromJob(job)
	guidance := fmt.Sprintf(fmtT, f.assertion, strings.ReplaceAll(input, "#", ""))
	var result struct {
		Asserted bool `json:"assertion_is_correct"`
	}
	err := llm.GenerateTypedJSON(job.GetContext(), f.client, guidance, f.model, jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"assertion_is_correct": {
				Type:        jsonschema.Boolean,
				Description: fmt.Sprintf("The following assertion is correct: %s", f.assertion),
			},
		},
		Required: []string{"assertion_is_correct"},
	}, &result)
	if err != nil {
		return false, err
	}

	if result.Asserted {
		return f.allowOnMatch, nil
	}
	return !f.allowOnMatch, nil
}

func (f *ClassifierFilter) IsTrigger() bool {
	return f.isTrigger
}

func ClassifierFilterConfigMeta() config.FieldGroup {
	return config.FieldGroup{
		Name:  FilterClassifier,
		Label: "Classifier Filter/Trigger",
		Fields: []config.Field{
			{Name: "name", Label: "Name", Type: "text", Required: true},
			{Name: "model", Label: "Model", Type: "text", Required: false, 
			  HelpText: "The LLM to use, usually a smaller one. Leave blank to use the same as the agent's"},
			{Name: "assertion", Label: "Assertion", Type: "text", Required: true,
			  HelpText: "Statement to match against e.g. 'The message is a question'"},
			{Name: "allow_on_match", Label: "Allow on Match", Type: "checkbox", Required: true},
			{Name: "is_trigger", Label: "Is Trigger", Type: "checkbox", Required: true},
		},
	}
}
