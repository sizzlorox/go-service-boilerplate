package models

import (
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Message string      `json:"message,omitempty" example:"Response Message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateResponse struct {
	InsertedID string `json:"insertedId" example:"5ff3fc0e00acd4328da25d92"`
}

type ValidationError struct {
	FailedField string `json:"failedField" example:"email"`
	Tag         string `json:"tag" example:"required"`
	Value       string `json:"value" example:""`
}

type Model struct {
	ID        string    `json:"id,omitempty" bson:"_id,omitempty" example:"5ff3fc0e00acd4328da25d92"`
	Name      string    `json:"name" bson:"name,omitempty" example:"Bob"`
	Email     string    `json:"email" bson:"email,omitempty" validate:"required,email" example:"bob@bob.com"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}

// https://pkg.go.dev/github.com/go-playground/validator
// ValidateStruct Validates if Struct is valid
func (m Model) ValidateStruct() []*ValidationError {
	var errors []*ValidationError
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

func (m Model) IsNil() bool {
	if m == (Model{}) {
		return true
	}
	switch reflect.TypeOf(m).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(m).IsNil()
	}
	return false
}
