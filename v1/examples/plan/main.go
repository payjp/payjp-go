package main

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
	"time"
)

func main() {
	payjpService := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	plans, _, _ := payjpService.Plan.List().Limit(1).Do()
	plan := plans[0]
	fmt.Println("  Id:", plan.ID)
	fmt.Println("  BillingDay:", plan.BillingDay)
	fmt.Println("  CreatedAt:", plan.CreatedAt.Format(time.RFC1123Z))
	fmt.Println("  Amount:", plan.Amount)
	fmt.Println("  Currency:", plan.Currency)
	fmt.Println("  Interval:", plan.Interval)
	fmt.Println("  LiveMode:", plan.LiveMode)
	fmt.Println("  Name:", plan.Name)
	fmt.Println("  TrialDays:", plan.TrialDays)
	fmt.Println("  Metadata:", plan.Metadata)
	err := plan.Update(payjp.Plan{
		Name: plan.Metadata["hoge"],
        Metadata: map[string]string{"hoge":plan.Name},
	})
	if err != nil {
		fmt.Println("plan update error")
	}
	fmt.Println("  Metadata:", plan.Metadata)
}
