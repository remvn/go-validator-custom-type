package main

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func makeValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// register func for custom struct, do this with every custom struct
	// you're going to need
	validate.RegisterCustomTypeFunc(validateValuer, sql.NullString{}, sql.NullInt64{})
	return validate
}

func validateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			// return the "real" value of custom struct
			// for example: if concrete type of driver.Valuer interface
			// is sql.NullString then val will be a string
			return val
		}
	}
	// return nil means this field is indeed "null".
	// field with tag `required` will fail the check
	return nil
}

func TestNullField(t *testing.T) {
	validate := makeValidator()

	type testCase struct {
		Name sql.NullString `validate:"required"`
	}
	test := &testCase{
		Name: sql.NullString{
			String: "Hello", // This is not empty on purpose,
			//	I will explain it later
			Valid: false, // Valid = false represent null
		},
	}

	err := validate.Struct(test)
	// check err is not nil
	assert.Error(t, err, "should return an error because Valid = false")
}

func TestFieldLength(t *testing.T) {
	validate := makeValidator()

	type testCase struct {
		Name sql.NullString `validate:"required,gt=10"`
	}
	test := &testCase{
		// This is a non-null string
		Name: sql.NullString{
			String: "hello",
			Valid:  true,
		},
	}

	err := validate.Struct(test)
	// check err is not nil
	assert.Error(t, err, "should return an error because length is less than 10")
}

func TestSimpleStruct(t *testing.T) {
	// check `Name` length is greater than 10
	type requestBody struct {
		Name string `validate:"required,gt=10"`
	}
	validate := makeValidator()

	body := &requestBody{
		Name: "short",
	}

	err := validate.Struct(body)
	assert.Error(t, err, "should return an error because name is less than 10")
}
