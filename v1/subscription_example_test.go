package payjp

import (
    "fmt"
    "testing"
)

func TestSubscriptionExample(t *testing.T) {
    service := New("sk_test_c62fade9d045b54cd76d7036", nil)
    subscriptions, _, err := service.Subscription.List().Do()
    if err != nil {
        t.Errorf("subscription list error")
    }
    subscription := subscriptions[0]
    fmt.Println("NextCyclePlan:", subscription.NextCyclePlan)
    id := subscription.ID
    plan, err := service.Plan.Create(Plan {
        Interval: "month",
        Currency: "jpy",
        Amount: 1000,
    })
    if err != nil {
        t.Errorf("plan create error")
        fmt.Println("err:", err)
    }
    setSubscr, err := service.Subscription.Update(id, Subscription{
        NextCyclePlanID: plan.ID,
    })
    if err != nil || setSubscr.NextCyclePlan.ID != plan.ID {
        t.Errorf("subscription update error with NextCyclePlan")
    }
    fmt.Println("NextCyclePlan:", setSubscr.NextCyclePlan)
    delSubscr, err := service.Subscription.Update(id, Subscription{
        NextCyclePlanID: "",
    })
    if err != nil || delSubscr.NextCyclePlan != nil {
        t.Errorf("subscription update error without NextCyclePlan")
    }
    fmt.Println("NextCyclePlan:", delSubscr.NextCyclePlan)
    if service.Plan.Delete(plan.ID) != nil {
        t.Errorf("plan delete error")
        fmt.Println("err:", err)
    }
}
