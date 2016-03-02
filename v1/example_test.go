package payjp_test

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
)

func ExampleTokenService() {
	pay := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	// 一度の支払いに使えるトークンを生成します。トークンIDはカード番号などの代わりに使用できます。
	fmt.Println("Create New Token")
	token, _ := pay.Token.Create(payjp.Card{
		Number:   "4242424242424242",
		CVC:      123,
		ExpMonth: 2,
		ExpYear:  2020,
	})
	fmt.Println("  Brand:", token.Card.Brand)
	fmt.Println("  CVC:", token.Card.CvcCheck)
	fmt.Println("  Last4:", token.Card.Last4)

	// 既存のトークン情報の取得
	fmt.Println("Find Existing Token")
	existingToken, _ := pay.Token.Retrieve(token.ID)
	fmt.Println("  Brand:", existingToken.Card.Brand)
	fmt.Println("  CVC:", existingToken.Card.CvcCheck)
	fmt.Println("  Last4:", existingToken.Card.Last4)
	// Output:
	// Create New Token
	//   Brand: Visa
	//   CVC: passed
	//   Last4: 4242
	// Find Existing Token
	//   Brand: Visa
	//   CVC: passed
	//   Last4: 4242
}

func ExampleChargeService() {
	pay := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	// 支払いをします
	charge, _ := pay.Charge.Create(3500, payjp.Charge{
		// 現在はjpyのみサポート
		Currency: "jpy",
		// カード情報、顧客ID、カードトークンのいずれかを指定
		Card: payjp.Card{
			Number:   "4242424242424242",
			CVC:      123,
			ExpMonth: 2,
			ExpYear:  2020,
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
	// Amount: 3500
	// Paid: true
}
