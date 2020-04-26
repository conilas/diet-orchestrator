package processors

import (
	"context"
	pb "diet-scheduler/be-test/pkg/food/v1"
	"log"
	conn "diet-scheduler/connections"
	db "diet-scheduler/database"
)

var KitchenToOrderStatusMap = map[pb.KitchenOrder_Status]pb.Order_Status{
	pb.KitchenOrder_PACKAGED: pb.Order_IN_FLIGHT,
}

func IsKitchenProcessed(status int) bool {
	kitchenStatus := pb.KitchenOrder_Status(status)

	return kitchenStatus == pb.KitchenOrder_PACKAGED
}

func KitchenToOrderStatusMapper(status int) int {
	return int(KitchenToOrderStatusMap[pb.KitchenOrder_Status(status)])
}

func CreateKitchenOrder(order string, ctx context.Context, client conn.ServiceClients) (string, error) {
	log.Printf("Sending order [%v] to Kitchen", order)

	kitchen, err := client.KitchenClient.CreateKitchenOrder(ctx, &pb.CreateKitchenOrderRequest{Kitchenorder: &pb.KitchenOrder{}})

	if err != nil {
		return "", err
	}

	return kitchen.Name, nil
}

func GetKitchenStatusByOrder(order string, ctx context.Context, clients conn.ServiceClients) (int, error) {
	relation, err := db.GetOrderRelations(clients.DatabaseClient, order)

	if err != nil {
		return int(pb.KitchenOrder_UNKNOWN), err
	}

	kitchen, err := clients.KitchenClient.GetKitchenOrder(ctx, &pb.GetKitchenOrderRequest{Name: relation.Kitchen})

	if err != nil {
		return int(pb.KitchenOrder_UNKNOWN), err
	}

	return int(kitchen.Status), nil
}

func ProcessKitchenToShipment(order string, ctx context.Context, client conn.ServiceClients, status int) {
	shipment, err := CreateShipmentOrder(order, ctx, client)

	if err != nil {
		log.Printf("Errored [%v]", err)
		return
	}

	log.Printf("Shipment order [%v] represents request [%v]", order, shipment)

	orderName, err := UpdateOrderStatus(order, ctx, client, pb.Order_Status(status))

	if err != nil {
		log.Printf("Errored [%v]", err)
		return
	}

	log.Printf("Order status [%v] updated!", orderName)

	db.SaveRelation(client.DatabaseClient, orderName, db.SHIPMENT, shipment)

}
