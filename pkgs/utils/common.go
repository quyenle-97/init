package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"runtime/debug"
	"strings"
)

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	// PostgreSQL unique violation check
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}

	// MySQL unique violation check
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 // Duplicate entry error
	}

	// Check error message as fallback method
	errMsg := err.Error()
	return strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "UNIQUE constraint failed")
}

func BindStruct[T any](m interface{}, v *T) error {
	if v == nil {
		return errors.New("destination pointer is nil")
	}
	jsonStr, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// Unmarshal the processed JSON into the target struct
	return json.Unmarshal(jsonStr, v)
}

func Validate(str interface{}) error {
	if str == nil {
		return errors.New("destination pointer is nil")
	}
	validate := validator.New()
	err := validate.Struct(str)
	if err != nil {
		return err
	}
	return nil
}

func Recovery() {
	if r := recover(); r != nil {
		fmt.Println(fmt.Sprintf("panic: %s - %s", r, string(debug.Stack())))
		fmt.Println("worker err: ", r)
	}
}

func Contains[T comparable](array []T, el T) bool {
	for _, a := range array {
		if a == el {
			return true
		}
	}
	return false
}

func Reverse[T any](s []T) []T {
	n := len(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func UrlWithPrefix(url string, base string) string {
	return fmt.Sprintf("%s%s", base, url)
}
