package main

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
)

func main() {
	payjpService := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	fmt.Println("See transfer summary")
	transfers, _, _ := payjpService.Transfer.List().Limit(1).Do()
	for i, transfer := range transfers {
		fmt.Printf("%d:\n", i)
		fmt.Println("  id:", transfer.ID)
		fmt.Println("  summary[charge_count]:", transfer.Summary.ChargeCount)
		fmt.Println("  summary[charge_fee]:", transfer.Summary.ChargeFee)
		fmt.Println("  summary[charge_gross]:", transfer.Summary.ChargeGross)
		fmt.Println("  summary[net]:", transfer.Summary.Net)
		fmt.Println("  summary[refund_amount]:", transfer.Summary.RefundAmount)
		fmt.Println("  summary[refund_count]:", transfer.Summary.RefundCount)
		fmt.Println("  summary[dispute_amount]:", transfer.Summary.DisputeAmount)
		fmt.Println("  summary[dispute_count]:", transfer.Summary.DisputeCount)
	}
}
