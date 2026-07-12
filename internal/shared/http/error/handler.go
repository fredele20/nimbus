package httperror

// func Handle(c *gin.Context, err error) {
// 	switch {
// 	case errors.Is(err, ErrUnauthorized):
// 		Error(
// 			c,
// 			http.StatusUnauthorized,
// 			"UNAUTHORIZED",
// 			err.Error(),
// 			// nil,
// 		)

// 	case errors.Is(err, ErrForbidden):
// 		Error(
// 			c,
// 			http.StatusForbidden,
// 			"FORBIDDEN",
// 			err.Error(),
// 			// nil,
// 		)

// 	case errors.Is(err, ErrNotFound):
// 		Error(
// 			c,
// 			http.StatusNotFound,
// 			"NOT_FOUND",
// 			err.Error(),
// 			// nil,
// 		)

// 	case errors.Is(err, ErrConflict):
// 		Error(
// 			c,
// 			http.StatusConflict,
// 			"CONFLICT",
// 			err.Error(),
// 			// nil,
// 		)

// 	default:
// 		Error(
// 			c,
// 			http.StatusInternalServerError,
// 			"INTERNAL_ERROR",
// 			"internal server error",
// 			// nil,
// 		)
// 	}
// }
