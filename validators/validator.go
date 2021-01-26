package validators

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"log"
)

var rfc3986Validator validator.Func = func(fl validator.FieldLevel) bool {

	return true
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("RFC3986", rfc3986Validator)
		if err != nil {
			log.Println("Error while registering rfc3986Validator to validations! error: " + err.Error())
		}
	}
}
