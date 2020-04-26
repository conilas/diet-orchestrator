package processors

import (
	"context"
	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
	db "diet-scheduler/database"
	"log"
)

var ShipmentToOrderStatusMap = map[pb.Shipment_Status]pb.Order_Status{
	pb.Shipment_DELIVERED: pb.Order_DELIVERED,
	pb.Shipment_REJECTED:  pb.Order_REJECTED,
}

func IsShipmentProcessed(status int) bool {
	shipmentStatus := pb.Shipment_Status(status)

	return shipmentStatus == pb.Shipment_DELIVERED || shipmentStatus == pb.Shipment_REJECTED
}

func ShipmentToOrderStatusMapper(status int) int {
	return int(ShipmentToOrderStatusMap[pb.Shipment_Status(status)])
}

func CreateShipmentOrder(order string, ctx context.Context, clients conn.ServiceClients) (string, error) {
	log.Printf("Sending order [%v] to Shipment", order)

	shipment, err := clients.ShipmentClient.CreateShipment(ctx, &pb.CreateShipmentRequest{Shipment: &pb.Shipment{}})

	if err != nil {
		return "", err
	}

	return shipment.Name, nil

}

func GetShipmentStatusByOrder(order string, ctx context.Context, clients conn.ServiceClients) (int, error) {
	relation, err := db.GetOrderRelations(clients.DatabaseClient, order)

	if err != nil {
		return int(pb.Shipment_UNKNOWN), err
	}

	shipment, err := clients.ShipmentClient.GetShipment(ctx, &pb.GetShipmentRequest{Name: relation.Shipment})

	if err != nil {
		return int(pb.Shipment_UNKNOWN), err
	}

	return int(shipment.Status), nil
}

func ProcessShipmentFinalization(order string, ctx context.Context, client conn.ServiceClients, newStatus int) {
	orderName, err := UpdateOrderStatus(order, ctx, client, pb.Order_Status(newStatus))

	if err != nil {
		log.Printf("Errored [%v]", err)
		return
	}

	log.Printf("[SHIPMENT] Order status [%v] updated to [%v]!", orderName, pb.Order_Status(newStatus))
}
