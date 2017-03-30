package main

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
)


func main() {
	pay := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	// 支払いをします
	charge, _ := pay.Charge.Create(3500, payjp.Charge{
		// 現在はjpyのみサポート
		Currency: "jpy",
		// カード情報、顧客ID、カードトークンのいずれかを指定
		Card: payjp.Card{
			Number: "4242424242424242",
			CVC:      "123",
			ExpMonth: "2",
			ExpYear:  "2020",
		},
		Capture: true,
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
}
