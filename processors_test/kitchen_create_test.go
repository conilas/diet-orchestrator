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

type CreateKitchenOrderSuite struct {
    suite.Suite
		clients conn.ServiceClients
		kitchenId string
    orderId string
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *CreateKitchenOrderSuite) SetupTest() {
	mockKitchenService := mock.NewMockKitchenServiceClient(suite.ctrl)
	kitchenName := "kitchens/AAAA"
	orderName := "orders/BBBB"

	mockKitchenService.EXPECT().CreateKitchenOrder(
		gomock.Any(),
    gomock.Any(),
	).Return(&pb.KitchenOrder{Name: kitchenName}, nil)

	suite.clients = conn.ServiceClients{
			KitchenClient: mockKitchenService,
	}

	suite.kitchenId = kitchenName
  suite.orderId = orderName
}

func (suite *CreateKitchenOrderSuite) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
}

func (suite *CreateKitchenOrderSuite) TestUpdateOrderToKitchen() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		name, err := processors.CreateKitchenOrder(suite.orderId, ctx, suite.clients)

		assert.Equal(suite.T(), err, nil, "No errors were supposed to be found after creating kitchen order")
		assert.Equal(suite.T(), name, suite.kitchenId, "Kitchen id was supposed to be equal to expected")
}

func TestCreateKitchenOrderSuite(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &CreateKitchenOrderSuite{ctrl: ctrl})
}

//Suite for when we are not able to create kitchen order

type UnableToCreateKitchenOrder struct {
    suite.Suite
		clients conn.ServiceClients
		kitchenId string
    orderId string
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *UnableToCreateKitchenOrder) SetupTest() {
	mockKitchenService := mock.NewMockKitchenServiceClient(suite.ctrl)
	kitchenName := "kitchens/AAAA"
	orderName := "orders/BBBB"

	mockKitchenService.EXPECT().CreateKitchenOrder(
		gomock.Any(),
    gomock.Any(),
	).Return(nil, errors.New("Unable to create kitchen order"))

	suite.clients = conn.ServiceClients{
			KitchenClient: mockKitchenService,
	}

	suite.kitchenId = kitchenName
  suite.orderId = orderName
}

func (suite *UnableToCreateKitchenOrder) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
}

func (suite *UnableToCreateKitchenOrder) TestUpdateOrderToKitchen() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		name, err := processors.CreateKitchenOrder(suite.orderId, ctx, suite.clients)

		assert.NotNil(suite.T(), err, "Should've returned an error")
		assert.Equal(suite.T(), name, "", "No ID for kitchen order because some error happened")
}

func TestUnableToCreateKitchenOrder(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &UnableToCreateKitchenOrder{ctrl: ctrl})
}
