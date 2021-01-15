package services

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sizzlorox/go-service-boilerplate/internal/datastore"
	"github.com/sizzlorox/go-service-boilerplate/internal/utils"
	"github.com/sizzlorox/go-service-boilerplate/internal/v1/models"
)

type Service interface {
	Get(page string, limit string) (resp ServiceResponse, err error)
	GetById(id string) (*ServiceResponse, error)
	Create(data *models.Model) (resp ServiceResponse, err error)
	Update(id string, data *models.Model) (*ServiceResponse, error)
	Delete(id string) (*ServiceResponse, error)
}

type service struct {
	r datastore.Repository
	u utils.Utils
}

type ServiceResponse struct {
	Status  int
	Message string      `json:"message" example:"Some Message"`
	Data    interface{} `json:"data,omitempty"`
}

/*
* CONSTRUCTOR
 */

func NewService(ds datastore.Repository, u utils.Utils) Service {
	return &service{r: ds, u: u}
}

/*
* PRIVATE
 */

func (s *service) mapPayload(res interface{}, response interface{}) error {
	// Mapping Response
	bRes, err := json.Marshal(res)
	if err != nil {
		log.Error(err)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	err = json.Unmarshal(bRes, &response)
	if err != nil {
		log.Error(err)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return nil
}

/*
* PUBLIC
 */

func (s *service) Get(page string, limit string) (resp ServiceResponse, err error) {
	// Build Query
	q := datastore.Query{From: "models"}

	// Convert params
	p, err := strconv.Atoi(page)
	if err != nil {
		return resp, err
	}

	// Validate Optional & Default params
	l := 30
	if len(limit) != 0 {
		l, err = strconv.Atoi(limit)
		if err != nil {
			return resp, err
		}
	}

	// Build Pagination Options
	pOpts := datastore.Pagination{
		Page:  p,
		Limit: l,
		Sort:  bson.M{"createdAt": 1},
	}

	// Datastore operation
	res, err := s.r.Paginate(q, pOpts)
	if err != nil {
		return resp, err
	}

	resp = ServiceResponse{
		Status:  fiber.StatusOK,
		Message: "Get Models Successful",
		Data:    res,
	}
	return resp, err
}

func (s *service) GetById(id string) (*ServiceResponse, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Build Query
	query := datastore.Query{
		Select: bson.M{"name": 1},
		Where:  bson.M{"_id": objectId},
		From:   "models",
	}

	// Datastore operation
	res, err := s.r.Find(query)
	if err != nil {
		return nil, err
	}
	if len((*res)) == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "Model Not Found")
	}

	return &ServiceResponse{
		Status:  fiber.StatusOK,
		Message: "Get Model by ID Successful",
		Data:    (*res)[0],
	}, err
}

func (s *service) Create(data *models.Model) (resp ServiceResponse, err error) {
	// Build Query
	query := datastore.Query{
		From: "models",
	}

	// Update Timestamp
	data.CreatedAt = time.Now().UTC()

	// Datastore operation
	res, err := s.r.Insert(query, data)
	if err != nil {
		err = s.u.ErrorWrapper(err)
		return resp, err
	}

	// Mapping Payload
	var payload models.CreateResponse
	err = s.mapPayload(res, &payload)
	if err != nil {
		return resp, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp = ServiceResponse{
		Status:  fiber.StatusCreated,
		Message: "Created Model Successfully",
		Data:    payload,
	}
	return resp, err
}

func (s *service) Update(id string, data *models.Model) (*ServiceResponse, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Build Query
	query := datastore.Query{
		Where: bson.M{"_id": objectId},
		From:  "models",
	}

	// Update Timestamp
	data.UpdatedAt = time.Now().UTC()

	// Datastore operation
	_, err = s.r.Update(query, bson.M{"$set": data})
	if err != nil {
		err = s.u.ErrorWrapper(err)
		return nil, err
	}

	return &ServiceResponse{
		Status:  fiber.StatusOK,
		Message: "Update Successful",
	}, err
}

func (s *service) Delete(id string) (*ServiceResponse, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Build Query
	query := datastore.Query{
		Where: bson.M{"_id": objectId},
		From:  "models",
	}

	// Datastore operation
	_, err = s.r.Delete(query)
	if err != nil {
		err = s.u.ErrorWrapper(err)
		return nil, err
	}

	return &ServiceResponse{
		Status:  fiber.StatusOK,
		Message: "Delete Successful",
	}, err
}
