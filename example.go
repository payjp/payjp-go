package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"v1/v1"
)

// smoke test を目的としたコード例

func display(resp interface{}) {
	b, _ := json.Marshal(&resp)
	var buf bytes.Buffer
	json.Indent(&buf, b, "", "  ")
	fmt.Println("RESPONSE:")
	fmt.Println(buf.String())
}

func abortIfNeeded(err error, msg string) {
	if err != nil {
		panic(msg + err.Error())
	}
	return
}

func createToken(pay *payjp.Service) string {
	resp, err := pay.Token.Create(payjp.Card{
		Name:     "A A",
		Number:   4242424242424242,
		ExpMonth: 10,
		ExpYear:  2028,
		CVC:      333,
	})
	abortIfNeeded(err, "CREATE TOKEN")
	fmt.Printf("USE: %s\n", resp.ID)
	return resp.ID
}

func accountApi(pay *payjp.Service, out bool) {
	// Account API の各エンドポイントを実行
	fmt.Println("=== ACCOUNT API ===")
	step := "*** RETRIEVE ACCOUNT ***"
	fmt.Println(step)
	resp, err := pay.Account.Retrieve()
	abortIfNeeded(err, step)
	if out {
		display(resp)
	}
}

func chargeApi(pay *payjp.Service) {
	// Charge API の各エンドポイントを実行
	fmt.Println("=== CHAREGE API ===")

	token := createToken(pay)
	step := "*** GATHERING CHARGE ***"
	resp, err := pay.Charge.Create(1000, payjp.Charge{
		Currency:    "jpy",
		CardToken:   token,
		Capture:     true,
		Description: "desc1",
		Metadata: map[string]string{
			"XXX": "YYY",
		},
	})
	abortIfNeeded(err, step)
	display(resp)

	step = "*** UPDATE CHARGE ***"
	resp, err = pay.Charge.Update(resp.ID, "new desc", map[string]string{"PPP": "ppp"})
	abortIfNeeded(err, step)
	display(resp)

	step = "*** WHOLE REFUND ***"
	resp, err = pay.Charge.Refund(resp.ID, "reason1", resp.Amount)
	abortIfNeeded(err, step)
	display(resp)

	step = "*** AUTH CHARGE ***"
	token = createToken(pay)
	resp, err = pay.Charge.Create(300, payjp.Charge{
		Currency:    "jpy",
		CardToken:   token,
		Capture:     false,
		Description: "desc capture",
	})
	abortIfNeeded(err, step)
	display(resp)

	step = "*** CAPTURE CHARGE ***"
	resp, err = pay.Charge.Capture(resp.ID, resp.Amount)
	abortIfNeeded(err, step)
	display(resp)

	step = "*** PIRTIAL REFUND ***"
	resp, err = pay.Charge.Refund(resp.ID, "paritial", resp.Amount-100)
	abortIfNeeded(err, step)
	display(resp)

	step = "*** RETRIEVE CHARGE ***"
	resp, err = pay.Charge.Retrieve(resp.ID)
	abortIfNeeded(err, step)
	display(resp)

	step = "*** GET CHARGES ***"
	charges, _, err := pay.Charge.List().Do()
	abortIfNeeded(err, step)
	display(charges)
}

func subscriptionAPI(pay *payjp.Service) {
	// Subscription API の各エンドポイントを実行
	// 便宜上、Plan API と Customer API もこの中で用いることとする
	fmt.Println("=== CUSTOMER API ===")
	step := "*** CREATE CUSTOMER ***"
	token := createToken(pay)
	cusresp, err := pay.Customer.Create(payjp.Customer{
		Email:       "this@is.test",
		Description: "cus",
		CardToken:   token,
		ID:          "cus1",
	})
	abortIfNeeded(err, step)
	display(cusresp)

	step = "*** UPDATE CUSTOMER ***"
	cusresp, err = pay.Customer.Update(cusresp.ID, payjp.Customer{
		Email: cusresp.Email + "2",
	})
	abortIfNeeded(err, step)
	display(cusresp)

	fmt.Println("=== PLAN API ===")
	step = "*** CREATE PLAN ***"
	plresp, err := pay.Plan.Create(payjp.Plan{
		Amount:   500,
		Currency: "jpy",
		Interval: "month",
		ID:       "plan1",
		Name:     "ppppppllllllaaaaannnnnn",
	})
	abortIfNeeded(err, step)
	display(plresp)

	step = "*** UPDATE PLAN ***"
	plresp, err = pay.Plan.Update(plresp.ID, "PPPPLLLLLNNNNNMMMM-XXX")
	abortIfNeeded(err, step)
	display(plresp)

	step = "RETRIEVE PLAN ***"
	plresp, err = pay.Plan.Retrieve(plresp.ID)
	abortIfNeeded(err, step)
	display(plresp)

	fmt.Println("=== SUBSCRIPTION API ===")
	step = "*** CREATE SUBSCRIPTION ***"
	subresp, err := pay.Subscription.Subscribe(cusresp.ID, payjp.Subscription{
		PlanID:     plresp.ID,
		TrialEndAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
	})
	abortIfNeeded(err, step)
	display(subresp)

	step = "*** UPDATE SUBSCRIPTION ***"
	subresp, err = pay.Subscription.Update(subresp.ID, payjp.Subscription{
		Metadata: map[string]string{
			"HHHH": "AAAAAA",
		},
	})
	abortIfNeeded(err, step)
	display(subresp)

	step = "*** RETRIEVE SUBSCRIPTION ***"
	subresp, err = pay.Subscription.Retrieve(cusresp.ID, subresp.ID)
	abortIfNeeded(err, step)
	display(subresp)

	step = "*** PAUSE SUBSCRIPTION ***"
	subresp, err = pay.Subscription.Pause(subresp.ID)
	abortIfNeeded(err, step)
	display(subresp)

	step = "*** RESUME SUBSCRIPTION ***"
	subresp, err = pay.Subscription.Resume(subresp.ID, payjp.Subscription{})
	abortIfNeeded(err, step)
	display(subresp)

	step = "*** CANCEL SUBSCRIPTION ***"
	subresp, err = pay.Subscription.Cancel(subresp.ID)
	abortIfNeeded(err, step)
	display(subresp)

	step = "*** DELETE SUBSCRIPTION ***"
	err = pay.Subscription.Delete(subresp.ID)
	abortIfNeeded(err, step)

	fmt.Println("=== PLAN API ===")
	step = "*** DELETE PLAN ***"
	err = pay.Plan.Delete(plresp.ID)
	abortIfNeeded(err, step)

	fmt.Println("=== CUSTOMER API ===")
	step = "*** DELETE CUSTOMER ***"
	err = pay.Customer.Delete(cusresp.ID)
	abortIfNeeded(err, step)
}

func concurrentAccountApi(pay *payjp.Service, i int, c chan int) {
	accountApi(pay, false)
	c <- i
}

func con(pay *payjp.Service, x int) {
	// 並列実行により意図的にテスト環境でレートリミットを発動させる
	c := make(chan int)
	for i := 0; i < x; i++ {
		fmt.Println(i)
		go concurrentAccountApi(pay, i, c)
	}
	fmt.Println(<-c)
}

func main() {
	key := "sk_test_c62fade9d045b54cd76d7036"
	var buf bytes.Buffer
	logger := log.New(&buf, "example_", log.Lshortfile)
	retryCount := 2
	retryConfig := payjp.RetryConfig{MaxCount: retryCount, InitialDelay: 2, MaxDelay: 32, Logger: logger}
	pay := payjp.New(key, nil, payjp.OptionRetryConfig(retryConfig))
	con(pay, 30)
	fmt.Println(buf.String())

	// accountApi(pay, true)
	// chargeApi(pay)
	// subscriptionAPI(pay)
}
