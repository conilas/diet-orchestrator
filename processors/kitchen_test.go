package processors

import (
	"testing"
)

func TestIsKitchenProcessed(t *testing.T) {
	var StatusTests = []struct {
		in   int
		out  bool
		desc string
	}{
		{int(pb.KitchenOrder_NEW), false, "New kitchen order is not processed"},
		{int(pb.KitchenOrder_PREPARATION), false, "'Preparation kitchen order' is not processed"},
		{int(pb.KitchenOrder_UNKNOWN), false, "Unknown kitchen order is not processed"},
		{int(pb.KitchenOrder_PACKAGED), true, "Packaged kitchen order is processed"},

		{10, false, "Unknown kitchen order is not processed"},
		{20, false, "Unknown kitchen order is not processed"},
		{100, false, "Unknown kitchen order is not processed"},
	}

	for _, tt := range StatusTests {
		t.Run(tt.desc, func(t *testing.T) {
			s := IsKitchenProcessed(tt.in)
			if s != tt.out {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}

func TestKitchenOrderToStatusMapper(t *testing.T) {
	var StatusTests = []struct {
		in   int
		out  int
		desc string
	}{
		{int(pb.KitchenOrder_PACKAGED), int(pb.Order_IN_FLIGHT), "Packaged on kitchen means order should go to in flight"},
		{int(pb.KitchenOrder_PREPARATION), int(pb.Order_UNKNOWN), "No mapping for PREPARATION"},
		{int(pb.KitchenOrder_UNKNOWN), int(pb.Order_UNKNOWN), "No mapping for UNKNOWN kitchen order status "},
		{int(pb.KitchenOrder_NEW), int(pb.Order_UNKNOWN), "No mapping for NEW kitchen order status"},
		{10, int(pb.Order_UNKNOWN), "No mapping for invalid status"},
		{20, int(pb.Order_UNKNOWN), "No mapping for invalid status"},
		{100, int(pb.Order_UNKNOWN), "No mapping for invalid status"},
	}

	for _, tt := range StatusTests {
		t.Run(tt.desc, func(t *testing.T) {
			s := KitchenToOrderStatusMapper(tt.in)
			if s != tt.out {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}
