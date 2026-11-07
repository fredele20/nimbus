package database

import (
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5/pgconn"
)

var UniqueIndexesWithValue = map[string]string{
    "users_email_key":        "email",
    "idx_users_phone_number": "phone",
}

type UniqueError struct {
    field string
}

func (e *UniqueError) Error() string {
    return fmt.Sprintf("%s already exists", e.field)
}

func CheckUniqueConstraintErr(err error) error {
    if err == nil {
        return nil
    }

    var pgErr *pgconn.PgError
    if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
        return err
    }

    field, exists := UniqueIndexesWithValue[pgErr.ConstraintName]
    if !exists {
        return err
    }

    return &UniqueError{field: field}
}

func IsIdxUniquenessError(err error) bool {
    var uniqueErr *UniqueError
    return errors.As(err, &uniqueErr)
}
