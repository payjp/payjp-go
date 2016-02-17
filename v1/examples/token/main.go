package main

import (
	"fmt"
	"github.com/payjp/payjp-go/v1"
)

func main() {
	payjpService := payjp.New("sk_test_c62fade9d045b54cd76d7036", nil)

	// create new token
	fmt.Println("Create New Token")
	token, _ := payjpService.Token.Create(payjp.Card{
		Number:   "4242424242424242",
		CVC:      123,
		ExpMonth: 2,
		ExpYear:  2020,
	})
	fmt.Println("  Id:", token.Id)
	fmt.Println("  Brand:", token.Card.Brand)
	fmt.Println("  CVC:", token.Card.CvcCheck)
	// Output:
	// Brand: Visa
	// CVC: passed

	// find existing token
	fmt.Println("Find Existing Token")
	existingToken, _ := payjpService.Token.Get(token.Id)
	fmt.Println("  Id:", existingToken.Id)
	fmt.Println("  Brand:", existingToken.Card.Brand)
	fmt.Println("  CVC:", existingToken.Card.CvcCheck)
	// Output:
	// Brand: Visa
	// CVC: passed
}
