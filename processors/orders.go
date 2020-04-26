package processors

import (
	"context"
	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
	db "diet-scheduler/database"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
	"log"
)

func UpdateOrderStatus(order string, ctx context.Context, client conn.ServiceClients, status pb.Order_Status) (string, error) {
	updateOrder, err := client.OrderClient.UpdateOrder(ctx, &pb.UpdateOrderRequest{Order: &pb.Order{Name: order, Status: status}, UpdateMask: &field_mask.FieldMask{Paths: []string{"status"}}})
	if err != nil {
		log.Printf("Errored [%v]", err)
		return "", err
	}
	return updateOrder.Name, nil
}

func ProcessOrderToKitchen(order string, ctx context.Context, client conn.ServiceClients) {

	kitchen, err := CreateKitchenOrder(order, ctx, client)

	if err != nil {
		log.Printf("Errored [%v]", err)
		return
	}
	log.Printf("Kitchen order [%v] represents request [%v]", order, kitchen)

	orderName, err := UpdateOrderStatus(order, ctx, client, pb.Order_PREPARATION)

	if err != nil {
		log.Printf("Errored [%v]", err)
		return
	}
	log.Printf("[ORDER] Order status [%v] updated to [%v]!", orderName, pb.Order_PREPARATION)

	db.SaveRelation(client.DatabaseClient, orderName, db.KITCHEN, kitchen)
}
