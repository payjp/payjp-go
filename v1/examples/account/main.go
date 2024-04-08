package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"github.com/payjp/payjp-go/v1"
)

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
	var buf bytes.Buffer
	logger := log.New(&buf, "example_", log.Lshortfile)
	retryConfig := payjp.Config{
		MaxCount: payjp.Int(5),
		// MaxDelay: payjp.Float(1),
		// InitialDelay: payjp.Float(0.001),
		Logger:   logger,
	}
	pay := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil, retryConfig)
	con(pay, 30)
	fmt.Println(buf.String())
}
