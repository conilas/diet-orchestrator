package database

import (
    "strings"
    "log"
    "context"
  	firestore "cloud.google.com/go/firestore"
)

type Relations struct {
  Kitchen string
  Order string
  Shipment string
}

type Phase int

const (
  collection = "relations"
  ordersPrefix = "orders/"
  empty = ""
)

const (
  KITCHEN Phase = 0
  SHIPMENT Phase = 1
)

var PhasesMap = map[Phase]string{
	KITCHEN: "kitchen",
  SHIPMENT: "shipment",
}

func GetOrderRelations(client firestore.Client, order string) (Relations, error){
	var c Relations

  ctx := context.Background()
  name := strings.ReplaceAll(order, ordersPrefix, empty)
  relation, err := client.Collection(collection).Doc(name).Get(ctx)

  if err != nil {
    return Relations{}, err
  }

  relation.DataTo(&c)

  return c, err
}

func SaveRelation(client firestore.Client, order string, phase Phase, value string) {
  ctx := context.Background()
  doc := strings.ReplaceAll(order, ordersPrefix, empty)

  _, err := client.Collection(collection).Doc(doc).Set(ctx, map[string]interface{}{
        PhasesMap[phase]: value,
  }, firestore.MergeAll)

  if err != nil {
        // Handle any errors in an appropriate way, such as returning them.
        log.Printf("An error has occurred: %s", err)
  }
}
