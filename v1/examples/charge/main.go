package main

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
)

func main() {
	pay := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	// 支払いをします

	// カードトークンを作成（サンプルのトークンは以下などで生成できます）
	// https://pay.jp/docs/checkout
	var tokenToCharge string = "tok_xxxxx"
	charge, _ := pay.Charge.Create(3500, payjp.Charge{
		// 現在はjpyのみサポート
		Currency:  "jpy",
		CardToken: tokenToCharge,
		Capture:   true,
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

	// 与信確保をします
	// 上記支払いをする際に使ったトークンとは別のものを指定してください。

	// カードトークンを作成
	var tokenToAuth string = "tok_yyyyy"

	authorizedCharge, _ := pay.Charge.Create(2800, payjp.Charge{
		// 現在はjpyのみサポート
		Currency:  "jpy",
		CardToken: tokenToAuth,
		Capture:   false,
		// 概要のテキストを設定できます
		Description: "Book: 'The Art of Community'",
		// 追加のメタデータを20件まで設定できます
		Metadata: map[string]string{
			"ISBN": "1449312063",
		},
	})
	fmt.Println("Amount:", authorizedCharge.Amount)
	fmt.Println("AmountRefunded:", authorizedCharge.AmountRefunded)
	fmt.Println("Paid:", authorizedCharge.Paid)
	fmt.Println("Captured:", authorizedCharge.Captured)
	// Output:
	// Amount: 2800
	// AmountRefunded: 0
	// Paid: true
	// Captured: false

	authorizedCharge.Capture(2000)
	capturedCharge, _ := pay.Charge.Retrieve(authorizedCharge.ID)
	fmt.Println("Amount:", capturedCharge.Amount)
	fmt.Println("AmountRefunded:", capturedCharge.AmountRefunded)
	fmt.Println("Capture:", capturedCharge.Captured)
	// Output:
	// Amount: 2800
	// AmountRefunded: 800
	// Captured: truea
}
