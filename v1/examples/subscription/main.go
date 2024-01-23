package main

import (
	"fmt"
	payjp "github.com/payjp/payjp-go/v1"
	"time"
)

func main() {
	s := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	fmt.Println("start Subscription scenario")
	// 支払い能力のある顧客を用意する。
	// このシナリオでは既に定期課金を持つ顧客を使う。
	status := payjp.SubscriptionActive
	subscriptions, _, err := s.Subscription.All(&payjp.SubscriptionListParams{
		ListParams: payjp.ListParams{
			Limit: payjp.Int(1),
		},
		Status: &status,
	})
	if err != nil {
		fmt.Println("cannot get subscription list:", err)
		return
	}
	fmt.Println("got subscription =", subscriptions[0].ID)
	fmt.Println("using customer =", subscriptions[0].Customer)

	// planを作成する
	plan, err := s.Plan.Create(payjp.Plan{
		Amount:   1000,
		Currency: "jpy",
		Interval: "month",
	})
	if err != nil {
		fmt.Println("cannot create plan:", err)
		return
	}
	fmt.Println("created plan =", plan)

	// subscriptionする
	subscription, err := s.Subscription.Subscribe(subscriptions[0].Customer, payjp.Subscription{
		PlanID: plan.ID,
		Metadata: map[string]string{
			"test": "created",
		},
	})
	if err != nil {
		fmt.Println("cannot create subscription:", err)
		return
	}
	fmt.Printf("%s is created with prorate %t\n", subscription.ID, subscription.Prorate)
	fmt.Println("detail =", subscription)

	// subscriptionを停止する
	err = subscription.Pause()
	if err != nil {
		fmt.Println("cannot pause subscription:", err)
		return
	}
	fmt.Printf("%s is %s\n", subscription.ID, subscription.Status)

	// subscriptionをトライアルで再開する
	nextYear := time.Now().AddDate(1, 0, 0)
	err = subscription.Resume(payjp.Subscription{
		TrialEndAt: nextYear,
	})
	if err != nil {
		fmt.Println("cannot resume subscription:", err)
		return
	}
	fmt.Printf("%s is %s until %s\n", subscription.ID, subscription.Status, subscription.TrialEndAt)

	// subscriptionをキャンセルする
	err = subscription.Cancel()
	if err != nil {
		fmt.Println("cannot pause subscription:", err)
		return
	}
	fmt.Printf("%s is %s\n", subscription.ID, subscription.Status)

	// subscriptionをトライアル解除して再開する
	err = subscription.Resume(payjp.Subscription{
		Prorate:   true,
		SkipTrial: true,
	})
	if err != nil {
		fmt.Println("cannot resume subscription:", err)
		return
	}
	fmt.Printf("%s is %s with prorate %t\n", subscription.ID, subscription.Status, subscription.Prorate)

	// subscriptionをリスト検索で見つける(plan+customerで一意)
	listedSubscriptions, hasMore, err := s.Subscription.All(&payjp.SubscriptionListParams{
		Plan:     payjp.String(plan.ID),
		Customer: payjp.String(subscription.Customer),
	})
	if err != nil || hasMore {
		fmt.Println("cannot search subscriptions:", err)
		return
	}
	fmt.Printf("%s is searched\n", listedSubscriptions[0].ID)

	// subscriptionを更新する
	// next_cycle_planを用意する
	nextCyclePlan, err := s.Plan.Create(payjp.Plan{
		Amount:   2000,
		Currency: "jpy",
		Interval: "month",
	})
	if err != nil {
		fmt.Println("cannot create plan:", err)
		return
	}
	fmt.Println("using next_cycle_plan =", nextCyclePlan)
	fmt.Printf("%s has metadata[test]=%s & next_cycle_plan=%s.\n",
		subscription.ID, subscription.Metadata["test"], subscription.NextCyclePlan)
	err = subscription.Update(payjp.Subscription{
		NextCyclePlanID: nextCyclePlan.ID,
		Metadata: map[string]string{
			"test": "updated",
		},
	})
	if err != nil {
		fmt.Println("subscription update error:", err)
		return
	}
	fmt.Printf("%s is updated for metadata[test]=%s & next_cycle_plan=%s.\n",
		subscription.ID, subscription.Metadata["test"], subscription.NextCyclePlan.ID)

	// subscriptionを削除する
	err = subscription.Delete()
	if err != nil {
		fmt.Println("subscription delete error:", err)
		return
	}
	fmt.Println("subscription deleted. so error occured by search.")

	// subscriptionを検索する
	_, err = s.Subscription.Retrieve(subscription.Customer, subscription.ID)
	fmt.Println(err)

	// 使われなくなったplanを削除する
	err = plan.Delete()
	if err != nil {
		fmt.Println(plan.ID+" delete error:", err)
		return
	}
	err = nextCyclePlan.Delete()
	if err != nil {
		fmt.Println(plan.ID+" delete error:", err)
		return
	}
}
