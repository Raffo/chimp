// +build novalidation
package validators

type DummyValidator struct{}

//Validate returns true if the passed interface is valid, false otherwise.
//If the interface cannot be passed, an error is returned.
func (*DummyValidator) Validate(input interface{}) (bool, error) {
	return true, nil
}

func NewDummyValidator() Validator {
	return &DummyValidator{}
}

func init() {
	New = NewDummyValidator
}
