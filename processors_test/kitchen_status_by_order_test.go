package processors_test

import (
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"diet-scheduler/processors"

	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
	db "diet-scheduler/database"
	mock "diet-scheduler/mocks"
)

type GetKitchenOrderByOrderNameSuite struct {
	suite.Suite
	clients        conn.ServiceClients
	kitchenId      string
	orderId        string
	expectedStatus pb.KitchenOrder_Status
	ctrl           *gomock.Controller
}

//sets up the mocks for this test
func (suite *GetKitchenOrderByOrderNameSuite) SetupTest() {
	mockKitchenService := mock.NewMockKitchenServiceClient(suite.ctrl)
	suite.kitchenId = "kitchens/AAA"
	suite.orderId = "orders/BBB"
	suite.expectedStatus = pb.KitchenOrder_PACKAGED

	mockKitchenService.EXPECT().GetKitchenOrder(
		gomock.Any(),
		&pb.GetKitchenOrderRequest{Name: suite.kitchenId},
	).Return(&pb.KitchenOrder{Name: suite.kitchenId, Status: suite.expectedStatus}, nil)

	setupDatabase(suite.orderId, suite.kitchenId)

	suite.clients = conn.ServiceClients{
		KitchenClient:  mockKitchenService,
		DatabaseClient: *conn.CreateFirestoreConnection(),
	}
}

func setupDatabase(order, kitchen string) {
	client := *conn.CreateFirestoreConnection()
	db.SaveRelation(client, order, db.KITCHEN, kitchen)
}

func (suite *GetKitchenOrderByOrderNameSuite) AfterTest(suiteName, testName string) {
	suite.ctrl.Finish()
	ctx := context.Background()
	cleanDoc := strings.ReplaceAll(suite.orderId, "orders/", "")

	log.Printf("Deleting %v", suite.orderId)

	_, err := suite.clients.DatabaseClient.Collection("relations").Doc(cleanDoc).Delete(ctx)

	if err != nil {
		log.Printf("Could not delete relations from firebase, %v", err)
	}
}

func (suite *GetKitchenOrderByOrderNameSuite) TestFetchKitchenStatusByOrderId() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	status, err := processors.GetKitchenStatusByOrder(suite.orderId, ctx, suite.clients)

	assert.Nil(suite.T(), err, "Err shoul've been nil")
	assert.Equal(suite.T(), status, int(suite.expectedStatus), "Expected status for kitchen order should have matched")
}

func TestGetKitchenOrderByOrderNameSuite(t *testing.T) {
	log.Printf("Starting %v", t.Name())
	ctrl := gomock.NewController(t)
	suite.Run(t, &GetKitchenOrderByOrderNameSuite{ctrl: ctrl})
}

// Suite for "not found" case

type GetKitchenOrderFailRequest struct {
	suite.Suite
	clients        conn.ServiceClients
	kitchenId      string
	orderId        string
	expectedStatus pb.KitchenOrder_Status
	ctrl           *gomock.Controller
}

//sets up the mocks for this test
func (suite *GetKitchenOrderFailRequest) SetupTest() {
	mockKitchenService := mock.NewMockKitchenServiceClient(suite.ctrl)
	suite.kitchenId = "kitchens/AAA"
	suite.orderId = "orders/BBB"
	suite.expectedStatus = pb.KitchenOrder_UNKNOWN

	mockKitchenService.EXPECT().GetKitchenOrder(
		gomock.Any(),
		&pb.GetKitchenOrderRequest{Name: suite.kitchenId},
	).Return(nil, errors.New("Unable to find kitchen order"))

	setupDatabase(suite.orderId, suite.kitchenId)

	suite.clients = conn.ServiceClients{
		KitchenClient:  mockKitchenService,
		DatabaseClient: *conn.CreateFirestoreConnection(),
	}
}

func (suite *GetKitchenOrderFailRequest) AfterTest(suiteName, testName string) {
	suite.ctrl.Finish()
	ctx := context.Background()
	cleanDoc := strings.ReplaceAll(suite.orderId, "orders/", "")

	log.Printf("Deleting %v", suite.orderId)

	_, err := suite.clients.DatabaseClient.Collection("relations").Doc(cleanDoc).Delete(ctx)

	if err != nil {
		log.Printf("Could not delete relations from firebase, %v", err)
	}
}

func (suite *GetKitchenOrderFailRequest) TestFetchKitchenStatusByOrderId() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	status, err := processors.GetKitchenStatusByOrder(suite.orderId, ctx, suite.clients)

	assert.NotNil(suite.T(), err, "There should've been an error when trying to fetch the kitchen order")
	assert.Equal(suite.T(), status, int(suite.expectedStatus), "Unknown kitchen order status - failed to fetch")
}

func TestGetKitchenOrderFailRequest(t *testing.T) {
	log.Printf("Starting %v", t.Name())
	ctrl := gomock.NewController(t)
	suite.Run(t, &GetKitchenOrderFailRequest{ctrl: ctrl})
}
