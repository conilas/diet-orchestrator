package processors_test

import (
  "errors"
	"testing"
  "strings"
	"context"
  "time"
  "log"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"

  "diet-scheduler/processors"

  pb "diet-scheduler/be-test/pkg/food/v1"
	mock "diet-scheduler/mocks"
	conn "diet-scheduler/connections"
)

type GetShipmentStatusByOrder struct {
    suite.Suite
		clients conn.ServiceClients
		shipmentId string
    orderId string
    expectedStatus pb.Shipment_Status
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *GetShipmentStatusByOrder) SetupTest() {
  mockShipmentService := mock.NewMockDroneServiceClient(suite.ctrl)
	suite.shipmentId = "shipments/AAA"
  suite.orderId = "orders/BBB"
  suite.expectedStatus = pb.Shipment_NEW

	mockShipmentService.EXPECT().GetShipment(
		gomock.Any(),
    &pb.GetShipmentRequest{Name: suite.shipmentId},
	).Return(&pb.Shipment{Name: suite.shipmentId, Status: suite.expectedStatus}, nil)

  setupDatabaseAllRelations(suite.orderId, "", suite.shipmentId)

	suite.clients = conn.ServiceClients{
			ShipmentClient: mockShipmentService,
			DatabaseClient: *conn.CreateFirestoreConnection(),
	}
}

func (suite *GetShipmentStatusByOrder) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
		ctx := context.Background()
		cleanDoc := strings.ReplaceAll(suite.orderId, "orders/", "")

		log.Printf("Deleting %v", suite.orderId)

		_, err := suite.clients.DatabaseClient.Collection("relations").Doc(cleanDoc).Delete(ctx)

		if err != nil {
			log.Printf("Could not delete relations from firebase, %v", err)
		}
}

func (suite *GetShipmentStatusByOrder) TestFetchKitchenStatusByOrderId() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		status, err := processors.GetShipmentStatusByOrder(suite.orderId, ctx, suite.clients)

		assert.Nil(suite.T(), err, "Err shoul've been nil")
		assert.Equal(suite.T(), status, int(suite.expectedStatus), "Expected status for kitchen order should have matched")
}

func TestGetShipmentStatusByOrder(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &GetShipmentStatusByOrder{ctrl: ctrl})
}


// Suite for "not found" case

type GetShipmentFailRequest struct {
    suite.Suite
		clients conn.ServiceClients
		shipmentId string
    orderId string
    expectedStatus pb.Shipment_Status
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *GetShipmentFailRequest) SetupTest() {
	mockShipmentService := mock.NewMockDroneServiceClient(suite.ctrl)
	suite.shipmentId = "kitchens/AAA"
  suite.orderId = "orders/BBB"
  suite.expectedStatus = pb.Shipment_UNKNOWN

	mockShipmentService.EXPECT().GetShipment(
		gomock.Any(),
    &pb.GetShipmentRequest{Name: suite.shipmentId},
	).Return(nil, errors.New("Unable to find kitchen order"))

  setupDatabaseAllRelations(suite.orderId, "", suite.shipmentId)

	suite.clients = conn.ServiceClients{
			ShipmentClient: mockShipmentService,
			DatabaseClient: *conn.CreateFirestoreConnection(),
	}
}

func (suite *GetShipmentFailRequest) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
		ctx := context.Background()
		cleanDoc := strings.ReplaceAll(suite.orderId, "orders/", "")

		log.Printf("Deleting %v", suite.orderId)

		_, err := suite.clients.DatabaseClient.Collection("relations").Doc(cleanDoc).Delete(ctx)

		if err != nil {
			log.Printf("Could not delete relations from firebase, %v", err)
		}
}

func (suite *GetShipmentFailRequest) TestFetchKitchenStatusByOrderId() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		status, err := processors.GetShipmentStatusByOrder(suite.orderId, ctx, suite.clients)

		assert.NotNil(suite.T(), err, "There should've been an error when trying to fetch the kitchen order")
		assert.Equal(suite.T(), status, int(suite.expectedStatus), "Unknown kitchen order status - failed to fetch")
}

func TestGetShipmentFailRequest(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &GetShipmentFailRequest{ctrl: ctrl})
}
