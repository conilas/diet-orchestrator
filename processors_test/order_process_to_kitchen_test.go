package processors_test

import (
	"testing"
	"context"
	"time"
	"log"
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

type ProcessOrderToKitchenSuite struct {
    suite.Suite
		clients conn.ServiceClients
		doc string
		kitchenId string
		ctrl *gomock.Controller
}

//sets up the mocks for this test
func (suite *ProcessOrderToKitchenSuite) SetupTest() {
	mockOrderService := mock.NewMockOrderServiceClient(suite.ctrl)
	mockKitchenService := mock.NewMockKitchenServiceClient(suite.ctrl)

	orderName := "orders/AAA"
	kitchenName := "orders/BBB"

  mockKitchenService.EXPECT().CreateKitchenOrder(
    gomock.Any(),
    gomock.Any(),
  ).Return(&pb.KitchenOrder{Name: kitchenName}, nil)

	mockOrderService.EXPECT().UpdateOrder(
		gomock.Any(),
		&pb.UpdateOrderRequest{Order: &pb.Order{Name: orderName, Status: pb.Order_PREPARATION}, UpdateMask: &field_mask.FieldMask{Paths: []string{"status"}}},
	).Return(&pb.Order{Name: orderName}, nil)

  setupDatabase(orderName, kitchenName)


	suite.clients = conn.ServiceClients{
			KitchenClient: mockKitchenService,
			OrderClient: mockOrderService,
			DatabaseClient: *conn.CreateFirestoreConnection(),
	}

	suite.doc = orderName
	suite.kitchenId = kitchenName
}

//cleans up the data from this test
//ideally we would mock firestore, but no good ways for such were found
func (suite *ProcessOrderToKitchenSuite) AfterTest(suiteName, testName string){
		suite.ctrl.Finish()
		ctx := context.Background()
		cleanDoc := strings.ReplaceAll(suite.doc, "orders/", "")

		log.Printf("Deleting %v", suite.doc)

		_, err := suite.clients.DatabaseClient.Collection("relations").Doc(cleanDoc).Delete(ctx)

		if err != nil {
			log.Printf("Could not delete relations from firebase, %v", err)
		}
}

func (suite *ProcessOrderToKitchenSuite) TestProcessKitchenToShipment() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		processors.ProcessOrderToKitchen(suite.doc, ctx, suite.clients)

		value, err := db.GetOrderRelations(suite.clients.DatabaseClient, suite.doc)

		assert.Equal(suite.T(), err, nil, "No errors were supposed to be found after fetching relations")
		assert.Equal(suite.T(), value.Kitchen, suite.kitchenId, "Kitchen id was supposed to be equal to expected")
}

func TestOrderToKitchenSuite(t *testing.T) {
    log.Printf("Starting %v", t.Name())
		ctrl := gomock.NewController(t)
		suite.Run(t, &ProcessOrderToKitchenSuite{ctrl: ctrl})
}
