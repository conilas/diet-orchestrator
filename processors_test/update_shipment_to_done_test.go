package processors_test

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
	db "diet-scheduler/database"
	mock "diet-scheduler/mocks"
	processors "diet-scheduler/processors"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
)

type SetShipmentToDoneSuite struct {
	suite.Suite
	clients    conn.ServiceClients
	doc        string
	kitchenId  string
	shipmentId string
	ctrl       *gomock.Controller
}

//sets up the mocks for this test
func (suite *SetShipmentToDoneSuite) SetupTest() {
	mockOrderService := mock.NewMockOrderServiceClient(suite.ctrl)
	mockKitchenService := mock.NewMockKitchenServiceClient(suite.ctrl)
	mockShipmentService := mock.NewMockDroneServiceClient(suite.ctrl)

	orderName := randOrdersName()
	kitchenName := randKitchenName()
	shipmentName := randShipmentName()

	var orders = []*pb.Order{
		&pb.Order{Name: orderName, Status: pb.Order_IN_FLIGHT},
	}

	mockOrderService.EXPECT().ListOrders(
		gomock.Any(),
		&pb.ListOrdersRequest{PageSize: 10},
	).Return(&pb.ListOrdersResponse{Orders: orders}, nil)

	mockShipmentService.EXPECT().GetShipment(
		gomock.Any(),
		&pb.GetShipmentRequest{Name: shipmentName},
	).Return(&pb.Shipment{Status: pb.Shipment_DELIVERED}, nil)

	mockOrderService.EXPECT().UpdateOrder(
		gomock.Any(),
		&pb.UpdateOrderRequest{Order: &pb.Order{Name: orderName, Status: pb.Order_DELIVERED}, UpdateMask: &field_mask.FieldMask{Paths: []string{"status"}}},
	).Return(&pb.Order{Name: orderName}, nil)

	setupDatabaseAllRelations(orderName, kitchenName, shipmentName)

	suite.clients = conn.ServiceClients{
		KitchenClient:  mockKitchenService,
		OrderClient:    mockOrderService,
		ShipmentClient: mockShipmentService,
		DatabaseClient: *conn.CreateFirestoreConnection(),
	}

	suite.doc = orderName
	suite.kitchenId = kitchenName
	suite.shipmentId = shipmentName
}

//cleans up the data from this test
//ideally we would mock firestore, but no good ways for such were found
func (suite *SetShipmentToDoneSuite) AfterTest(suiteName, testName string) {
	suite.ctrl.Finish()
	ctx := context.Background()
	cleanDoc := strings.ReplaceAll(suite.doc, "orders/", "")

	log.Printf("Deleting %v", suite.doc)

	_, err := suite.clients.DatabaseClient.Collection("relations").Doc(cleanDoc).Delete(ctx)

	if err != nil {
		log.Printf("Could not delete relations from firebase, %v", err)
	}
}

func (suite *SetShipmentToDoneSuite) TestUpdateOrderToKitchen() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := suite.clients.OrderClient.ListOrders(ctx, &pb.ListOrdersRequest{PageSize: 10})
	log.Printf("Reply %v %v", r, err)
	processors.ProcessSingleOrder(*r.Orders[0], ctx, suite.clients)

	value, err := db.GetOrderRelations(suite.clients.DatabaseClient, suite.doc)

	assert.Equal(suite.T(), err, nil, "No errors were supposed to be found after fetching relations")
	assert.Equal(suite.T(), value.Kitchen, suite.kitchenId, "Kitchen id was supposed to be equal to expected")
	assert.Equal(suite.T(), value.Shipment, suite.shipmentId, "Kitchen id was supposed to be equal to expected")
}

func TestShipmentToDoneSuite(t *testing.T) {
	log.Printf("Starting %v", t.Name())
	ctrl := gomock.NewController(t)
	suite.Run(t, &SetShipmentToDoneSuite{ctrl: ctrl})
}
