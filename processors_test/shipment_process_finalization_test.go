package processors_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
	mock "diet-scheduler/mocks"
	processors "diet-scheduler/processors"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
)

type ProcessShipmentFinalizationSuite struct {
	suite.Suite
	clients    conn.ServiceClients
	doc        string
	shipmentId string
	ctrl       *gomock.Controller
}

//sets up the mocks for this test
func (suite *ProcessShipmentFinalizationSuite) SetupTest() {
	mockOrderService := mock.NewMockOrderServiceClient(suite.ctrl)

	orderName := "orders/AAA"
	shipmentName := "orders/CCC"

	mockOrderService.EXPECT().UpdateOrder(
		gomock.Any(),
		&pb.UpdateOrderRequest{Order: &pb.Order{Name: orderName, Status: pb.Order_DELIVERED}, UpdateMask: &field_mask.FieldMask{Paths: []string{"status"}}},
	).Return(&pb.Order{Name: orderName}, nil)

	suite.clients = conn.ServiceClients{
		OrderClient:    mockOrderService,
		DatabaseClient: *conn.CreateFirestoreConnection(),
	}

	suite.doc = orderName
	suite.shipmentId = shipmentName
}

func (suite *ProcessShipmentFinalizationSuite) AfterTest(suiteName, testName string) {
	suite.ctrl.Finish()
}

func (suite *ProcessShipmentFinalizationSuite) TestShipmentFinalizationSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	processors.ProcessShipmentFinalization(suite.doc, ctx, suite.clients, int(pb.Order_DELIVERED))
}

func TestShipmentFinalizationSuite(t *testing.T) {
	log.Printf("Starting %v", t.Name())
	ctrl := gomock.NewController(t)
	suite.Run(t, &ProcessShipmentFinalizationSuite{ctrl: ctrl})
}
