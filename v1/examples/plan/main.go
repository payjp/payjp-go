package main

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
	"time"
)

func main() {
	payjpService := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	fmt.Println("Getting existing plan")
	plans, hasMore, _ := payjpService.Plan.List().Limit(10).Offset(10).Do()
	fmt.Println("hasMore:", hasMore)
	for i, plan := range plans {
		fmt.Printf("%d:\n", i)
		fmt.Println("  Id:", plan.ID)
		fmt.Println("  BillingDay:", plan.BillingDay)
		fmt.Println("  CreatedAt:", plan.CreatedAt.Format(time.RFC1123Z))
		fmt.Println("  Amount:", plan.Amount)
		fmt.Println("  Currency:", plan.Currency)
		fmt.Println("  Interval:", plan.Interval)
		fmt.Println("  LiveMode:", plan.LiveMode)
		fmt.Println("  Name:", plan.Name)
		fmt.Println("  TrialDays:", plan.TrialDays)
	}
}
