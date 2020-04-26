package processors

import (
	"context"
	pb "diet-scheduler/be-test/pkg/food/v1"
	conn "diet-scheduler/connections"
	"log"
	"time"
)

type StatusFetcher func(string, context.Context, conn.ServiceClients) (int, error)
type IntPredicate func(int) bool
type Processor func(string, context.Context, conn.ServiceClients, int)
type StatusMapper func(int) int
type ComposedProcessor func(string, context.Context, conn.ServiceClients)

func ProcessNextStep(fetcher StatusFetcher, validator IntPredicate, mapper StatusMapper, process Processor) ComposedProcessor {
	return func(order string, ctx context.Context, client conn.ServiceClients) {
		status, err := fetcher(order, ctx, client)

		if err != nil {
			log.Printf("Errored [%v]", err)
			return
		}

		if validator(status) {
			process(order, ctx, client, mapper(status))
		}

	}
}

var orderProcessment ComposedProcessor = ProcessOrderToKitchen
var kitchenProcessment ComposedProcessor = ProcessNextStep(GetKitchenStatusByOrder, IsKitchenProcessed, KitchenToOrderStatusMapper, ProcessKitchenToShipment)
var deliveryProcess ComposedProcessor = ProcessNextStep(GetShipmentStatusByOrder, IsShipmentProcessed, ShipmentToOrderStatusMapper, ProcessShipmentFinalization)

var StatusToProcessFunction = map[pb.Order_Status] ComposedProcessor{
	pb.Order_NEW: orderProcessment,
	pb.Order_PREPARATION: kitchenProcessment,
	pb.Order_IN_FLIGHT: deliveryProcess,
}

func ProcessSingleOrder(order pb.Order, ctx context.Context, client conn.ServiceClients) {
	log.Printf("Order [%v] with current status [%v]", order.Name, order.Status)

	processFn := StatusToProcessFunction[order.Status]

	if processFn != nil {
		processFn(order.Name, ctx, client)
	}
}

func ProcessAllOrders(orders pb.ListOrdersResponse, ctx context.Context, client conn.ServiceClients) {
	for _, order := range orders.Orders {
		go ProcessSingleOrder(*order, ctx, client)
	}

	if orders.NextPageToken != "" {
		newOrders, err := client.OrderClient.ListOrders(ctx, &pb.ListOrdersRequest{PageToken: orders.NextPageToken, PageSize: 10})
		if err != nil {
			log.Printf("Invalid request with token [%v]. Error: %v", orders.NextPageToken, err)
			return
		}
		ProcessAllOrders(*newOrders, ctx, client)
	}
}

func ScheduledJob(server, certificate string) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clients := conn.CreateConnections(server, certificate)
	orders, err := clients.OrderClient.ListOrders(ctx, &pb.ListOrdersRequest{PageSize: 10})
	if err != nil {
		log.Printf("Could not fetch orders. Wait untill next try. Err: [%v]", err)
		return
	}
	ProcessAllOrders(*orders, ctx, clients)
}
