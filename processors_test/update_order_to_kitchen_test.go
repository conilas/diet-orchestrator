package processors_test

import (
	"testing"
	"context"
	"time"
	"log"
  "math/rand"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"

	processors "diet-scheduler/processors"
	db "diet-scheduler/database"
	mock "diet-scheduler/mocks"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
		rand.Seed(time.Now().Unix())

    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func randKitchenName() string{
	return "kitchens/" + randString(10)
}

func randOrdersName() string{
		return "orders/" + randString(10)
}

func randShipmentName() string{
		return "shipments/" + randString(10)
}

type SetOrderToKitchenSuite struct {
    suite.Suite
		clients conn.ServiceClients
		doc string
		kitchenId string
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *SetOrderToKitchenSuite) SetupTest() {
	mockOrderService := mock.NewMockOrderServiceClient(suite.ctrl)
	mockKitchenService := mock.NewMockKitchenServiceClient(suite.ctrl)
	mockShipmentService := mock.NewMockDroneServiceClient(suite.ctrl)

	orderName := randOrdersName()
	kitchenName := randKitchenName()

	var orders = []*pb.Order {
		&pb.Order{Name: orderName, Status: pb.Order_NEW},
	}

	mockOrderService.EXPECT().ListOrders(
		gomock.Any(),
		&pb.ListOrdersRequest{PageSize: 10},
	).Return(&pb.ListOrdersResponse{Orders: orders}, nil)

	mockKitchenService.EXPECT().CreateKitchenOrder(
		gomock.Any(),
		gomock.Any(),
	).Return(&pb.KitchenOrder{Name: kitchenName}, nil)

	mockOrderService.EXPECT().UpdateOrder(
		gomock.Any(),
		&pb.UpdateOrderRequest{Order: &pb.Order{Name: orderName, Status: pb.Order_PREPARATION}, UpdateMask: &field_mask.FieldMask{Paths: []string{"status"}}},
	).Return(&pb.Order{Name: orderName}, nil)

	suite.clients = conn.ServiceClients{
			KitchenClient: mockKitchenService,
			OrderClient: mockOrderService,
			ShipmentClient: mockShipmentService,
			DatabaseClient: *conn.CreateFirestoreConnection(),
	}

	suite.doc = orderName
	suite.kitchenId = kitchenName
}

//cleans up the data from this test
//ideally we would mock firestore, but no good ways for such were found
func (suite *SetOrderToKitchenSuite) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
		ctx := context.Background()
		cleanDoc := strings.ReplaceAll(suite.doc, "orders/", "")

		log.Printf("Deleting %v", suite.doc)

		_, err := suite.clients.DatabaseClient.Collection("relations").Doc(cleanDoc).Delete(ctx)

		if err != nil {
			log.Printf("Could not delete relations from firebase, %v", err)
		}
}

func (suite *SetOrderToKitchenSuite) TestUpdateOrderToKitchen() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := suite.clients.OrderClient.ListOrders(ctx, 	&pb.ListOrdersRequest{PageSize: 10})
		log.Printf("Reply %v %v", r, err)
		processors.ProcessSingleOrder(*r.Orders[0], ctx, suite.clients)

		value, err := db.GetOrderRelations(suite.clients.DatabaseClient, suite.doc)

		assert.Equal(suite.T(), err, nil, "No errors were supposed to be found after fetching relations")
		assert.Equal(suite.T(), value.Kitchen, suite.kitchenId, "Kitchen id was supposed to be equal to expected")
}

func TestUpdateOrderToKitchenSingleOrderProcessing(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &SetOrderToKitchenSuite{ctrl: ctrl})
}