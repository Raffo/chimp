// +build zalandoValidation

package validators

import (
	. "github.com/zalando-techmonkeys/chimp/types"
	"strings"
)

type ZalandoValidator struct{}

//Validate returns true if the passed interface is valid, false otherwise.
//If the interface cannot be passed, an error is returned.
func (*ZalandoValidator) Validate(input interface{}) (bool, error) {
	//validate to chimp create/update data structure
	//TODO this validation is only basic and needs improvement
	//validating pierone only url
	dr := input.(DeployRequest)
	if !strings.HasPrefix(dr.ImageURL, "pierone.stups.zalan.do") {
		return false, nil
	}
	return true, nil
}

func NewZalandoValidator() Validator {
	return &ZalandoValidator{}
}

func init() {
	New = NewZalandoValidator
}
