package processors

import (
	"testing"

  pb "diet-scheduler/be-test/pkg/food/v1"
)

func TestIsShipmentProcessed(t *testing.T) {
	var StatusTests = []struct {
		in   int
		out  bool
		desc string
	}{
		{int(pb.Shipment_DELIVERED), true, "Delivery means processed"},
		{int(pb.Shipment_REJECTED), true, "Rejected means processed"},
    {int(pb.Shipment_NEW), false, "New shipment is not processed"},
    {int(pb.Shipment_UNKNOWN), false, "Unknown shipment is not processed"},

		{10, false, "Unknown shipment is not processed"},
		{20, false, "Unknown shipment is not processed"},
		{100, false, "Unknown shipment is not processed"},
	}

	for _, tt := range StatusTests {
		t.Run(tt.desc, func(t *testing.T) {
			s := IsShipmentProcessed(tt.in)
			if s != tt.out {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}

func TestShipmentToOrderStatusMapper(t *testing.T) {
	var StatusTests = []struct {
		in   int
		out  int
		desc string
	}{
		{int(pb.Shipment_DELIVERED), int(pb.Order_DELIVERED), "Delivered shipmnet means delivered order"},
		{int(pb.Shipment_REJECTED), int(pb.Order_REJECTED), "Rejected shipmnet means rejected order"},
    {int(pb.Shipment_NEW), int(pb.Order_UNKNOWN), "No mapping for NEW shipment status "},
    {int(pb.Shipment_UNKNOWN), int(pb.Order_UNKNOWN), "No mapping for UNKNOWN shipment status"},

    {10, int(pb.Order_UNKNOWN), "No mapping for invalid status"},
    {20, int(pb.Order_UNKNOWN), "No mapping for invalid status"},
    {100, int(pb.Order_UNKNOWN), "No mapping for invalid status"},
	}

	for _, tt := range StatusTests {
		t.Run(tt.desc, func(t *testing.T) {
			s := ShipmentToOrderStatusMapper(tt.in)
			if s != tt.out {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}
