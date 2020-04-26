package connections

import (
	firestore "cloud.google.com/go/firestore"
	pb "diet-scheduler/be-test/pkg/food/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

type ServiceClients struct {
	KitchenClient  pb.KitchenServiceClient
	OrderClient    pb.OrderServiceClient
	ShipmentClient pb.DroneServiceClient
	DatabaseClient firestore.Client
}

func CreateConnections(address, certificate string) ServiceClients {
	creds, err := credentials.NewClientTLSFromFile(certificate, "")

	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
	}

	conn, _ := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	return ServiceClients{OrderClient: pb.NewOrderServiceClient(conn),
		KitchenClient:  pb.NewKitchenServiceClient(conn),
		DatabaseClient: *CreateFirestoreConnection(),
		ShipmentClient: pb.NewDroneServiceClient(conn)}
}
