package processors_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
	mock "diet-scheduler/mocks"
	"diet-scheduler/processors"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
)

type UpdateOrderStatusSuite struct {
	suite.Suite
	clients     conn.ServiceClients
	doc         string
	afterStatus pb.Order_Status
	ctrl        *gomock.Controller
}

//sets up the mocks for this test
func (suite *UpdateOrderStatusSuite) SetupTest() {
	mockOrderService := mock.NewMockOrderServiceClient(suite.ctrl)
	orderName := "orders/AAA"
	suite.afterStatus = pb.Order_IN_FLIGHT

	mockOrderService.EXPECT().UpdateOrder(
		gomock.Any(),
		&pb.UpdateOrderRequest{Order: &pb.Order{Name: orderName, Status: suite.afterStatus}, UpdateMask: &field_mask.FieldMask{Paths: []string{"status"}}},
	).Return(&pb.Order{Name: orderName}, nil)

	mockOrderService.EXPECT().GetOrder(
		gomock.Any(),
		&pb.GetOrderRequest{Name: orderName},
	).Return(&pb.Order{Name: orderName, Status: suite.afterStatus}, nil)

	suite.clients = conn.ServiceClients{
		OrderClient: mockOrderService,
	}

	suite.doc = orderName
}

//cleans up the data from this test
//ideally we would mock firestore, but no good ways for such were found
func (suite *UpdateOrderStatusSuite) AfterTest(suiteName, testName string) {
	suite.ctrl.Finish()
}

func (suite *UpdateOrderStatusSuite) TestProcessKitchenToShipment() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	doc, err := processors.UpdateOrderStatus(suite.doc, ctx, suite.clients, suite.afterStatus)
	value, _ := suite.clients.OrderClient.GetOrder(ctx, &pb.GetOrderRequest{Name: suite.doc})

	assert.Equal(suite.T(), err, nil, "No errors were supposed to be found after fetching relations")
	assert.Equal(suite.T(), doc, suite.doc, "Id was not supposed to change")
	assert.Equal(suite.T(), value.Status, suite.afterStatus, "Status was supposed to be changed")
}

func TestUpdateStausOrder(t *testing.T) {
	log.Printf("Starting %v", t.Name())
	ctrl := gomock.NewController(t)
	suite.Run(t, &UpdateOrderStatusSuite{ctrl: ctrl})
}
