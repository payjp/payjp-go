// PAY.JP for Golang
//
// PAY.JP API Docs: https://pay.jp/docs/api/
//
// Installation
//
// You can download Pay.jp Golang SDK by the following command:
//
//   $ go get github.com/payjp/payjp-go/v1
//
// How To Use
//
// Entry point of this package is payjp.New():
//
//   pay := payjp.New("api key", nil)
//
// pay has several children like Token, Customer, Plan, Subscription etc:
//
//   customer := pay.Customer.Get("customer ID")
//
// Have fun!

package payjp
