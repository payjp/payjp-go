package main

import (
	"fmt"
	"./v1"
)


func main() {
	pay := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	// 支払いをします
	charge, _ := pay.Charge.Create(3500, payjp.Charge{
		// 現在はjpyのみサポート
		Currency: "jpy",
		// カード情報、顧客ID、カードトークンのいずれかを指定
    CustomerID: "cus_cb6df76d665aa2296ea7100edc9b",
		Capture: false,
		// 概要のテキストを設定できます
		Description: "Book: 'The Art of Community'",
		// 追加のメタデータを20件まで設定できます
		Metadata: map[string]string{
			"ISBN": "1449312063",
		},
	})
	fmt.Println("Amount:", charge.Amount)
	fmt.Println("Paid:", charge.Paid)
	// Output:
	// Paid: true
  one, _ := pay.Charge.Retrieve(charge.ID)
  one.Capture()
}
