package validators

//Validator is the interface used for validation related functionalities
type Validator interface {
	//Validate returns true if the passed interface is valid, false otherwise.
	//If the interface cannot be passed, an error is returned.
	Validate(input interface{}) (bool, error)
}

type validatorFactory func() Validator

var New validatorFactory
