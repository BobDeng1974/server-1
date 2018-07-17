package helpers

import (
	"errors"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

// PropertyFixYaml help with issues of unknown configs' data structure.
// For example default
// field	interface{}
// with data
// color:
//   r : 120
// 	 g : 120
//   b : 120
// will be un-marshaled as map[interface{}] interface{}.
// Which prevents normal deep compare.
func PropertyFixYaml(x interface{}, p enums.Property) (interface{}, error) {
	if nil == x {
		return x, nil
	}

	switch p {
	case enums.PropColor:
		return convertProperty(x, &common.Color{})
	case enums.PropOn:
		r, ok := x.(bool)
		if !ok {
			return nil, errors.New("error converting bool")
		}

		return r, nil
	case enums.PropBrightness:
		return convertValueProperty(x, &common.Percent{})
	}

	return x, nil
}

// PropertyDeepEqual uses some extended rules for different common types.
// For example we don't care about scenes updates, so it's always true.
func PropertyDeepEqual(x, y interface{}, p enums.Property) bool {
	switch p {
	case enums.PropScenes:
		// No updates for scenes
		return true
	default:
		return cmp.Equal(x, y)
	}
}

// CommandPropertyFixYaml fixes properties similar to PropertyFixYaml method.
func CommandPropertyFixYaml(x interface{}, c enums.Command) (interface{}, error) {
	if nil == x {
		return x, nil
	}

	switch c {
	case enums.CmdOn, enums.CmdOff, enums.CmdToggle:
		return nil, nil
	case enums.CmdSetBrightness:
		return convertValueProperty(x, &common.Percent{})
	case enums.CmdSetTransitionTime:
		return convertValueProperty(x, &common.Int{})
	case enums.CmdSetColor:
		return convertProperty(x, &common.Color{})
	}

	return x, nil
}

// Converts to default value-based property.
func convertValueProperty(from, to interface{}) (interface{}, error) {
	wrap := map[string]interface{}{"value": from}
	return convertProperty(wrap, to)
}

// Converts property to target type.
func convertProperty(from, to interface{}) (interface{}, error) {
	data, err := yaml.Marshal(from)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, to)
	return to, err
}