package validation

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(c *gin.Context, data any) error {
	if bindErr := c.ShouldBindJSON(data); bindErr != nil {
		return bindErr
	}

	err := validate.Struct(data)
	if err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
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
