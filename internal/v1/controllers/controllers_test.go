package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/models"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ControllerSuite struct {
	suite.Suite
}

type MockService struct {
	mock.Mock
	services.Service
}

/*
	MOCKS
*/
func (m *MockService) Create(model *models.Model) (services.ServiceResponse, error) {
	args := m.Called(model)
	return args.Get(0).(services.ServiceResponse), args.Error(1)
}

func (suite *ControllerSuite) SetupTest() {}

func (suite *ControllerSuite) TestCreate() {
	t := suite.T()

	tests := []struct {
		description string

		// Test input
		route          string
		payload        models.Model
		mockedResponse services.ServiceResponse

		// Expected output
		expectedError bool
		expectedCode  int
		expectedBody  string
	}{
		{
			description: "[Create] Success",
			route:       "/api/v1/create",
			payload: models.Model{
				Name:  "test",
				Email: "test@test.com",
			},
			mockedResponse: services.ServiceResponse{
				Status:  fiber.StatusCreated,
				Message: "Created Model Successfully",
			},
			expectedError: false,
			expectedCode:  fiber.StatusCreated,
			expectedBody:  "{\"Status\":201,\"message\":\"Created Model Successfully\"}",
		},
		{
			description: "[Create] Missing Required Field Email",
			route:       "/api/v1/create",
			payload: models.Model{
				Name: "test",
			},
			mockedResponse: services.ServiceResponse{
				Status:  fiber.StatusCreated,
				Message: "Created Model Successfully",
			},
			expectedError: true,
			expectedCode:  fiber.StatusBadRequest,
			expectedBody:  "[{\"failedField\":\"Model.Email\",\"tag\":\"required\",\"value\":\"\"}]",
		},
		{
			description: "[Create] Missing Required Field Name",
			route:       "/api/v1/create",
			payload: models.Model{
				Email: "test@test.com",
			},
			mockedResponse: services.ServiceResponse{
				Status:  fiber.StatusCreated,
				Message: "Created Model Successfully",
			},
			expectedError: true,
			expectedCode:  fiber.StatusBadRequest,
			expectedBody:  "[{\"failedField\":\"Model.Name\",\"tag\":\"required\",\"value\":\"\"}]",
		},
	}

	// create an instance of our test object
	mockService := new(MockService)

	// Initialize stuff
	controller := NewController(mockService)

	app := fiber.New()
	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Post("/create", controller.Create)

	for _, test := range tests {
		// Mock Service call
		mockService.On("Create", &test.payload).Return(test.mockedResponse, nil)

		// Setup Payload
		jsonBytes, err := json.Marshal(test.payload)
		if err != nil {
			assert.Nil(t, err)
		}
		contentBuffer := bytes.NewBuffer(jsonBytes)

		// Setup Request
		req, _ := http.NewRequest("POST", "/api/v1/create", contentBuffer)
		req.Header.Set("Content-Type", "application/json")

		res, err := app.Test(req, -1)

		// Asserts
		assert.Equal(t, test.expectedCode, res.StatusCode, test.description)
		body, err := ioutil.ReadAll(res.Body)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		assert.Nilf(t, err, test.description)
		assert.Equalf(t, test.expectedBody, string(body), test.description)
	}
}

func TestRunControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerSuite))
}
