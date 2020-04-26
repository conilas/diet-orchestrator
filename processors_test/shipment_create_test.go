package processors_test

import (
  "errors"
	"testing"
	"context"
  "time"
  "log"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"


  processors "diet-scheduler/processors"
  pb "diet-scheduler/be-test/pkg/food/v1"
	mock "diet-scheduler/mocks"
	conn "diet-scheduler/connections"
)

type CreateShipmentSuite struct {
    suite.Suite
		clients conn.ServiceClients
		shipmentId string
    orderId string
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *CreateShipmentSuite) SetupTest() {
	mockShipmentService := mock.NewMockDroneServiceClient(suite.ctrl)
	shipmentName := "shipment/AAAA"
	orderName := "orders/BBBB"

	mockShipmentService.EXPECT().CreateShipment(
    gomock.Any(),
    gomock.Any(),
	).Return(&pb.Shipment{Name: shipmentName}, nil)

	suite.clients = conn.ServiceClients{
			ShipmentClient: mockShipmentService,
	}

	suite.shipmentId = shipmentName
  suite.orderId = orderName
}

func (suite *CreateShipmentSuite) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
}

func (suite *CreateShipmentSuite) TestCreateShipment() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		name, err := processors.CreateShipmentOrder(suite.orderId, ctx, suite.clients)

		assert.Equal(suite.T(), err, nil, "No errors were supposed to be found after creating kitchen order")
		assert.Equal(suite.T(), name, suite.shipmentId, "Kitchen id was supposed to be equal to expected")
}

func TestCreateShipmentSuite(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &CreateShipmentSuite{ctrl: ctrl})
}


//Suite for when we are not able to create kitchen order

type UnableToCreateShipmentOrderSuite struct {
    suite.Suite
		clients conn.ServiceClients
		shipmentId string
    orderId string
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *UnableToCreateShipmentOrderSuite) SetupTest() {
	mockShipmentService := mock.NewMockDroneServiceClient(suite.ctrl)
	shipmentName := "shipments/AAAA"
	orderName := "orders/BBBB"

	mockShipmentService.EXPECT().CreateShipment(
		gomock.Any(),
    gomock.Any(),
	).Return(nil, errors.New("Unable to create kitchen order"))

	suite.clients = conn.ServiceClients{
			ShipmentClient: mockShipmentService,
	}

	suite.shipmentId = shipmentName
  suite.orderId = orderName
}

func (suite *UnableToCreateShipmentOrderSuite) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
}

func (suite *UnableToCreateShipmentOrderSuite) TestUpdateOrderToKitchen() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		name, err := processors.CreateShipmentOrder(suite.orderId, ctx, suite.clients)

		assert.NotNil(suite.T(), err, "Should've returned an error")
		assert.Equal(suite.T(), name, "", "No ID for kitchen order because some error happened")
}

func TestUnableToCreateShipmentOrderSuite(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &UnableToCreateShipmentOrderSuite{ctrl: ctrl})
}
