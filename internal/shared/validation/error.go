package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// Tell the validator to use the 'json' tag name for field errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
}

func Validate(c *gin.Context, data any) error {
	if bindErr := c.ShouldBindJSON(data); bindErr != nil {
		return bindErr
	}

	err := validate.Struct(data)
	if err != nil {
		if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
			return fmt.Errorf(
				"Validation failed for field '%s', invalid value '%v'",
				ve[0].Field(),
				ve[0].Value(),
			)
		}
		return err
	}

	return nil
}
