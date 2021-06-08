package main

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
)

func main() {
	service := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)
	subscriptions, _, err := service.Subscription.List().Do()
	if err != nil {
		fmt.Println("subscription list error")
	}
	subscription := subscriptions[0]
	fmt.Println("NextCyclePlan:", subscription.NextCyclePlan)
	id := subscription.ID

	plan, err := service.Plan.Create(payjp.Plan{
		Interval: "month",
		Currency: "jpy",
		Amount:   1000,
	})
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("Plan:", plan)

	setSubscr, err := service.Subscription.Update(id, payjp.Subscription{
		NextCyclePlanID: plan.ID,
	})
	fmt.Println("NextCyclePlan:", setSubscr.NextCyclePlan)

	delSubscr, err := service.Subscription.Update(id, payjp.Subscription{
		NextCyclePlanID: "",
	})
	fmt.Println("NextCyclePlan:", delSubscr.NextCyclePlan)
	if service.Plan.Delete(plan.ID) != nil {
		fmt.Println("err:", err)
	}
}
