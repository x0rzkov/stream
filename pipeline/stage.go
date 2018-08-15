package pipeline

import (
	"errors"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	)

var (
	exists = struct{}{}
)

// switch to this mapper style in future?
type MapperAlt interface {
	Apply(inputScope data.Scope) (map[string]*data.Attribute, error)
}

type Stage struct {
	act      activity.Activity
	settings map[string]*data.Attribute
	inputs   *InputValues

	outputAttrs map[string]*data.Attribute

	// do we need outputMapper instead?
	promote map[string]struct{} //promote these outputs to the pipeline
}

type StageConfig struct {
	*activity.Config

	Promotions    []string           `json:"addToPipeline,omitempty"`
}

func NewStage(config *StageConfig) (*Stage, error) {

	if config.Ref == "" {
		return nil, errors.New("Activity not specified for Stage")
	}

	act := activity.Get(config.Ref)
	if act == nil {
		return nil, errors.New("Unsupported Activity:" + config.Ref)
	}

	stage := &Stage{}
	stage.act = act

	if len(config.Settings) > 0 {
		stage.settings = make(map[string]*data.Attribute, len(config.Settings))

		for name, value := range config.Settings {

			attr := act.Metadata().Settings[name]

			if attr != nil {
				//todo handle error
				stage.settings[name], _ = data.NewAttribute(name, attr.Type(), resolveSettingValue(name, value))
			}
		}
	}

	inputAttrs := config.InputAttrs

	if len(inputAttrs) > 0 {

		var err error
		stage.inputs, err = NewInputValues(act.Metadata().Input, GetDataResolver(), inputAttrs)

		if err != nil {
			return nil, err
		}
	}

	outputAttrs := config.OutputAttrs

	if len(outputAttrs) > 0 {

		stage.outputAttrs = make(map[string]*data.Attribute, len(outputAttrs))

		for name, value := range outputAttrs {

			attr := act.Metadata().Output[name]

			if attr != nil {
				//todo handle error
				stage.outputAttrs[name], _ = data.NewAttribute(name, attr.Type(), value)
			}
		}
	}

	if len(config.Promotions) > 0 {
		stage.promote = make(map[string]struct{})

		for _, value := range config.Promotions {
			stage.promote[value] = exists
		}
	}

	return stage, nil
}

func resolveSettingValue(setting string, value interface{}) interface{} {

	strVal, ok := value.(string)

	if ok && len(strVal) > 0 && strVal[0] == '$' {
		v, err := data.GetBasicResolver().Resolve(strVal, nil)

		if err == nil {

			logger.Debugf("Resolved setting [%s: %s] to : %v", setting, value, v)
			return v
		}
	}

	return value
}
