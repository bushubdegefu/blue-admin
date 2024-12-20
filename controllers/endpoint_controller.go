package controllers

import (
	"net/http"
	"strconv"

	"blue-admin.com/common"
	"blue-admin.com/models"
	"blue-admin.com/observe"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetEndpointis a function to get a Endpoints by ID
// @Summary Get Endpoints
// @Description Get Endpoints
// @Tags Endpoints
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security Refresh
// @Param page query int true "page"
// @Param size query int true "page size"
// @Success 200 {object} common.ResponsePagination{data=[]models.EndpointGet}
// @Failure 404 {object} common.ResponseHTTP{}
// @Router /endpoint [get]
func GetEndpoints(contx *fiber.Ctx) error {

	//  Getting tracer context
	ctx := contx.Locals("tracer")
	tracer, _ := ctx.(*observe.RouteTracer)

	//  Getting Database connection
	db, _ := contx.Locals("db").(*gorm.DB)

	//  parsing Query Prameters
	Page, _ := strconv.Atoi(contx.Query("page"))
	Limit, _ := strconv.Atoi(contx.Query("size"))
	//  checking if query parameters  are correct
	if Page == 0 || Limit == 0 {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: "Not Allowed, Bad request",
			Data:    nil,
		})
	}

	//  querying result with pagination using gorm function
	result, err := common.PaginationPureModel(db, models.Endpoint{}, []models.Endpoint{}, uint(Page), uint(Limit), tracer.Tracer)
	if err != nil {
		return contx.Status(http.StatusInternalServerError).JSON(common.ResponseHTTP{
			Success: false,
			Message: "Failed to get all Endpoint.",
			Data:    "something",
		})
	}

	// returning result if all the above completed successfully
	return contx.Status(http.StatusOK).JSON(result)
}

// GetEndpointByID is a function to get a Endpoints by ID
// @Summary Get Endpoint by ID
// @Description Get endpoint by ID
// @Tags Endpoints
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param endpoint_id path int true "Endpoint ID"
// @Success 200 {object} common.ResponseHTTP{data=models.EndpointGet}
// @Failure 404 {object} common.ResponseHTTP{}
// @Router /endpoint/{endpoint_id} [get]
func GetEndpointByID(contx *fiber.Ctx) error {

	// Starting tracer context and tracer
	ctx := contx.Locals("tracer")
	tracer, _ := ctx.(*observe.RouteTracer)

	//  Getting Database connection
	db, _ := contx.Locals("db").(*gorm.DB)

	//  parsing Query Prameters
	id, err := strconv.Atoi(contx.Params("endpoint_id"))
	if err != nil {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// Preparing and querying database using Gorm
	var endpoints_get models.EndpointGet
	var endpoints models.Endpoint
	if res := db.WithContext(tracer.Tracer).Model(&models.Endpoint{}).Preload(clause.Associations).Where("id = ?", id).First(&endpoints); res.Error != nil {
		return contx.Status(http.StatusNotFound).JSON(common.ResponseHTTP{
			Success: false,
			Message: res.Error.Error(),
			Data:    nil,
		})
	}

	// filtering response data according to filtered defined struct
	mapstructure.Decode(endpoints, &endpoints_get)

	//  Finally returing response if All the above compeleted successfully
	return contx.Status(http.StatusOK).JSON(common.ResponseHTTP{
		Success: true,
		Message: "Success got one endpoint.",
		Data:    &endpoints_get,
	})
}

// GetAppEndpointUUID is a function to get a App enpoints by UUID
// @Summary Get App Endpoints by UUID
// @Description Get app endpoints by UUID
// @Tags Endpoints
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int true "page"
// @Param size query int true "page size"
// @Param app_uuid path string true "App UUID"
// @Success 200 {object} common.ResponseHTTP{data=[]models.EndpointPut}
// @Failure 404 {object} common.ResponseHTTP{}
// @Router /appendpointuuid/{app_uuid} [get]
func GetAppEndpointsAllUUID(contx *fiber.Ctx) error {

	// Starting tracer context and tracer
	ctx := contx.Locals("tracer")
	tracer, _ := ctx.(*observe.RouteTracer)

	//  Getting Database connection
	db, _ := contx.Locals("db").(*gorm.DB)

	//  parsing Query Prameters
	Page, _ := strconv.Atoi(contx.Query("page"))
	Limit, _ := strconv.Atoi(contx.Query("size"))
	//  checking if query parameters  are correct
	if Page == 0 || Limit == 0 {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: "Not Allowed, Bad request",
			Data:    nil,
		})
	}

	//  parsing Query Prameters
	uuid := contx.Params("app_uuid")
	if uuid == "" {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: "No uuid",
			Data:    nil,
		})
	}

	// Preparing and querying database using Gorm
	//getting total count first
	var total_counter int64
	count_string := `select distinct endpoints.id from endpoints inner join features on
							endpoints.feature_id = features.id
							inner join roles on features.role_id= roles.id
							inner join apps on roles.app_id == apps.id
							where apps.uuid = ? ORDER BY endpoints.id;`
	if res := db.WithContext(tracer.Tracer).Raw(count_string, uuid, Limit, Page).Count(&total_counter); res.Error != nil {
		return contx.Status(http.StatusNotFound).JSON(common.ResponseHTTP{
			Success: false,
			Message: res.Error.Error(),
			Data:    nil,
		})
	}

	var endpoints []models.EndpointPut
	// select apps.id as appID, roles.id, roles.name, roles.description,roles.active from roles inner join apps on roles.app_id == apps.id where apps.uuid =="0191c74f-d039-71c6-a3be-66e2571a9cf1" ORDER BY roles.id;
	query_string := `select distinct endpoints.id,endpoints.name, endpoints.method,endpoints.route_path,endpoints.description from endpoints inner join features on
							endpoints.feature_id = features.id
							inner join roles on features.role_id= roles.id
							inner join apps on roles.app_id == apps.id
							where apps.uuid = ? ORDER BY endpoints.id LIMIT ? OFFSET ?;`

	if res := db.WithContext(tracer.Tracer).Raw(query_string, uuid, Limit, Page-1).Scan(&endpoints); res.Error != nil {
		return contx.Status(http.StatusNotFound).JSON(common.ResponseHTTP{
			Success: false,
			Message: res.Error.Error(),
			Data:    nil,
		})
	}

	//  Finally returing response if All the above compeleted successfully
	return contx.Status(http.StatusOK).JSON(common.ResponseHTTP{
		Success: true,
		Message: "Success got one app.",
		Total:   uint(total_counter),
		Page:    uint(Page),
		Size:    uint(Limit),
		Data:    &endpoints,
	})
}

// Add Endpoint to data
// @Summary Add a new Endpoint
// @Description Add Endpoint
// @Tags Endpoints
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param endpoint body models.EndpointPost true "Add Endpoint"
// @Success 200 {object} common.ResponseHTTP{data=models.EndpointPost}
// @Failure 400 {object} common.ResponseHTTP{}
// @Failure 500 {object} common.ResponseHTTP{}
// @Router /endpoint [post]
func PostEndpoint(contx *fiber.Ctx) error {

	// Starting tracer context and tracer
	ctx := contx.Locals("tracer")
	tracer, _ := ctx.(*observe.RouteTracer)

	// Getting Database Connection
	db, _ := contx.Locals("db").(*gorm.DB)

	// validator initialization
	validate := validator.New()

	//validating post data
	posted_endpoint := new(models.EndpointPost)

	//first parse request data
	if err := contx.BodyParser(&posted_endpoint); err != nil {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// then validate structure
	if err := validate.Struct(posted_endpoint); err != nil {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	//  initiate -> endpoint
	endpoint := new(models.Endpoint)
	endpoint.Name = posted_endpoint.Name
	endpoint.Description = posted_endpoint.Description
	endpoint.Method = posted_endpoint.Method
	endpoint.RoutePath = posted_endpoint.RoutePath

	//  start transaction to database
	tx := db.WithContext(tracer.Tracer).Begin()

	// add  data using transaction if values are valid
	if err := tx.Create(&endpoint).Error; err != nil {
		tx.Rollback()
		return contx.Status(http.StatusInternalServerError).JSON(common.ResponseHTTP{
			Success: false,
			Message: "Endpoint Creation Failed",
			Data:    err,
		})
	}

	// close transaction
	tx.Commit()

	// return data if transaction is sucessfull
	return contx.Status(http.StatusOK).JSON(common.ResponseHTTP{
		Success: true,
		Message: "Endpoint created successfully.",
		Data:    endpoint,
	})
}

// Patch Endpoint to data
// @Summary Patch Endpoint
// @Description Patch Endpoint
// @Tags Endpoints
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param endpoint body models.EndpointPost true "Patch Endpoint"
// @Param endpoint_id path int true "Endpoint ID"
// @Success 200 {object} common.ResponseHTTP{data=models.EndpointPost}
// @Failure 400 {object} common.ResponseHTTP{}
// @Failure 500 {object} common.ResponseHTTP{}
// @Router /endpoint/{endpoint_id} [patch]
func PatchEndpoint(contx *fiber.Ctx) error {

	// Starting tracer context and tracer
	ctx := contx.Locals("tracer")
	tracer, _ := ctx.(*observe.RouteTracer)

	// Get database connection
	db, _ := contx.Locals("db").(*gorm.DB)

	//  initialize data validator
	validate := validator.New()

	// validate path params
	id, err := strconv.Atoi(contx.Params("endpoint_id"))
	if err != nil {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// validate data struct
	patch_endpoint := new(models.EndpointPatch)
	if err := contx.BodyParser(&patch_endpoint); err != nil {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// then validating
	if err := validate.Struct(patch_endpoint); err != nil {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// startng update transaction
	var endpoint models.Endpoint
	tx := db.WithContext(tracer.Tracer).Begin()

	// Check if the record exists
	if err := db.WithContext(tracer.Tracer).Where("id = ?", id).First(&endpoint).Error; err != nil {
		// If there's an unexpected error, return an internal server error response
		tx.Rollback()
		return contx.Status(http.StatusNotFound).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// Update the record
	if err := db.WithContext(tracer.Tracer).Model(&endpoint).UpdateColumns(*patch_endpoint).Error; err != nil {
		tx.Rollback()
		return contx.Status(http.StatusInternalServerError).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// Return  success response
	return contx.Status(http.StatusOK).JSON(common.ResponseHTTP{
		Success: true,
		Message: "Endpoint updated successfully.",
		Data:    endpoint,
	})
}

// DeleteEndpoints function removes a endpoint by ID
// @Summary Remove Endpoint by ID
// @Description Remove endpoint by ID
// @Tags Endpoints
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param endpoint_id path int true "Endpoint ID"
// @Success 200 {object} common.ResponseHTTP{}
// @Failure 404 {object} common.ResponseHTTP{}
// @Failure 503 {object} common.ResponseHTTP{}
// @Router /endpoint/{endpoint_id} [delete]
func DeleteEndpoint(contx *fiber.Ctx) error {

	// Starting tracer context and tracer
	ctx := contx.Locals("tracer")
	tracer, _ := ctx.(*observe.RouteTracer)

	// Getting Database connection
	db, _ := contx.Locals("db").(*gorm.DB)

	// get deleted endpoint attributes to return
	var endpoint models.Endpoint

	// validate path params
	id, err := strconv.Atoi(contx.Params("endpoint_id"))
	if err != nil {
		return contx.Status(http.StatusBadRequest).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// perform delete operation if the object exists
	tx := db.WithContext(tracer.Tracer).Begin()

	// first getting endpoint and checking if it exists
	if err := db.Where("id = ?", id).First(&endpoint).Error; err != nil {
		tx.Rollback()
		return contx.Status(http.StatusNotFound).JSON(common.ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	// Delete the endpoint
	if id > 78 {
		if err := db.Delete(&endpoint).Error; err != nil {
			tx.Rollback()
			return contx.Status(http.StatusInternalServerError).JSON(common.ResponseHTTP{
				Success: false,
				Message: "Error deleting endpoint",
				Data:    nil,
			})
		}
	}

	// Commit the transaction
	tx.Commit()

	// Return success respons
	return contx.Status(http.StatusOK).JSON(common.ResponseHTTP{
		Success: true,
		Message: "Endpoint deleted successfully.",
		Data:    endpoint,
	})
}

// ################################################################
// Relationship Based Endpoints
// ################################################################

type EndPointDropDown struct {
	ID   uint   `validate:"required" json:"id"`
	Name string `validate:"required" json:"name"`
}

// Get EndPoint Dropdown only active roles
// @Summary Get EndPointDropDown
// @Description Get EndPointDropDown
// @Tags EndPoints
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} common.ResponseHTTP{data=[]EndPointDropDown}
// @Failure 404 {object} common.ResponseHTTP{}
// @Router /endpointdrop [get]
func GetDropEndPoints(contx *fiber.Ctx) error {
	//  Getting tracer context
	ctx := contx.Locals("tracer")
	tracer, _ := ctx.(*observe.RouteTracer)

	//  Getting Database connection
	db, _ := contx.Locals("db").(*gorm.DB)

	var features_drop []EndPointDropDown
	if res := db.WithContext(tracer.Tracer).Model(&models.Endpoint{}).Find(&features_drop); res.Error != nil {
		return contx.Status(http.StatusServiceUnavailable).JSON(common.ResponseHTTP{
			Success: false,
			Message: res.Error.Error(),
			Data:    nil,
		})
	}

	return contx.Status(http.StatusOK).JSON(common.ResponseHTTP{
		Success: true,
		Message: "Success got one role.",
		Data:    &features_drop,
	})
}
