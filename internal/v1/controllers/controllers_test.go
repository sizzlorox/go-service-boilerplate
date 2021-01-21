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

type TestCase struct {
	description string

	// Test input
	method         string
	route          string
	payload        models.Model
	mockedResponse services.ServiceResponse
	mockedError    error

	// Expected output
	expectedError bool
	expectedCode  int
	expectedBody  string
}

func (tc TestCase) CaseRunner(app *fiber.App) (*http.Response, []byte, error) {
	// Setup Payload
	jsonBytes, _ := json.Marshal(tc.payload)
	contentBuffer := bytes.NewBuffer(jsonBytes)

	// Setup Request
	req, _ := http.NewRequest(tc.method, tc.route, contentBuffer)
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req, -1)

	// Asserts
	body, err := ioutil.ReadAll(res.Body)
	return res, body, err
}

/*
	MOCKS
*/
// TODO: Maybe move somewhere else or smthing?
func (m *MockService) Create(model *models.Model) (services.ServiceResponse, error) {
	args := m.Called(model)
	return args.Get(0).(services.ServiceResponse), args.Error(1)
}

func (m *MockService) Update(id string, model *models.Model) (services.ServiceResponse, error) {
	args := m.Called(id, model)
	return args.Get(0).(services.ServiceResponse), args.Error(1)
}

/*
	TESTS
*/

// func (suite *ControllerSuite) SetupTest() {}

func (suite *ControllerSuite) TestCreate() {
	t := suite.T()

	tests := []TestCase{
		{
			description: "[Create] Success",
			method:      "PUT",
			route:       "/api/v1/create",
			payload: models.Model{
				Name:  "test",
				Email: "test@test.com",
			},
			mockedResponse: services.ServiceResponse{
				Status:  fiber.StatusCreated,
				Message: "Created Model Successfully",
			},
			mockedError:   nil,
			expectedError: false,
			expectedCode:  fiber.StatusCreated,
			expectedBody:  "{\"Status\":201,\"message\":\"Created Model Successfully\"}",
		},
		{
			description: "[Create] Missing Required Field Email",
			method:      "PUT",
			route:       "/api/v1/create",
			payload: models.Model{
				Name: "test",
			},
			mockedResponse: services.ServiceResponse{
				Status:  fiber.StatusCreated,
				Message: "Created Model Successfully",
			},
			mockedError:   nil,
			expectedError: true,
			expectedCode:  fiber.StatusBadRequest,
			expectedBody:  "[{\"failedField\":\"Model.Email\",\"tag\":\"required\",\"value\":\"\"}]",
		},
		{
			description: "[Create] Missing Required Field Name",
			method:      "PUT",
			route:       "/api/v1/create",
			payload: models.Model{
				Email: "test@test.com",
			},
			mockedResponse: services.ServiceResponse{
				Status:  fiber.StatusCreated,
				Message: "Created Model Successfully",
			},
			mockedError:   nil,
			expectedError: true,
			expectedCode:  fiber.StatusBadRequest,
			expectedBody:  "[{\"failedField\":\"Model.Name\",\"tag\":\"required\",\"value\":\"\"}]",
		},
	}

	// create an instance of our test object
	mockService := new(MockService)

	// Initialize stuff
	controller := NewController(mockService)

	// TODO: Move errorhandler somewhere else and use it here to validate mockedError
	app := fiber.New()
	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Put("/create", controller.Create)

	for _, test := range tests {
		// Mock Service call
		mockService.On("Create", &test.payload).Return(test.mockedResponse, test.mockedError)
		res, body, err := test.CaseRunner(app)

		// Asserts
		assert.Equal(t, test.expectedCode, res.StatusCode, test.description)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		assert.Nilf(t, err, test.description)
		assert.Equalf(t, test.expectedBody, string(body), test.description)
	}
}

func (suite *ControllerSuite) TestUpdate() {
	t := suite.T()

	tests := []TestCase{
		{
			description: "[Update] Success",
			method:      "POST",
			route:       "/api/v1/mockid/update",
			payload: models.Model{
				Name:  "test",
				Email: "test@test.com",
			},
			mockedResponse: services.ServiceResponse{
				Status:  fiber.StatusOK,
				Message: "Update Successful",
			},
			mockedError:   nil,
			expectedError: false,
			expectedCode:  fiber.StatusOK,
			expectedBody:  "{\"Status\":200,\"message\":\"Update Successful\"}",
		},
		{
			description:    "[Update] Empty Payload",
			method:         "POST",
			route:          "/api/v1/mockid/update",
			payload:        models.Model{},
			mockedResponse: services.ServiceResponse{}, // We wont reach this point in this test case
			expectedError:  true,
			expectedCode:   fiber.StatusBadRequest,
			expectedBody:   "You require at least one field to update",
		},
	}

	// create an instance of our test object
	mockService := new(MockService)

	// Initialize stuff
	controller := NewController(mockService)

	app := fiber.New()
	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Post("/:id/update", controller.Update)

	for _, test := range tests {
		// Mock Service call
		mockService.On("Update", "mockid", &test.payload).Return(test.mockedResponse, test.mockedError)
		res, body, err := test.CaseRunner(app)

		// Asserts
		assert.Equal(t, test.expectedCode, res.StatusCode, test.description)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		assert.Nilf(t, err, test.description)
		assert.Equalf(t, test.expectedBody, string(body), test.description)
	}
}

func TestRunControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerSuite))
}
