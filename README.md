# PAY.JP for Go

[![Build Status](https://github.com/payjp/payjp-go/actions/workflows/build-test.yml/badge.svg?branch=master)](https://github.com/payjp/payjp-go/actions)

## このVersionは移行期間中のためデフォルトブランチとは分離されています。デフォルトブランチにマージ予定ですので早めの検証をお願いします。

## Installation

go.modにbetaブランチのコードを記載ください。

```
$cat go.mod
...
require (
    github.com/payjp/payjp-go beta
)

$go mod tidy
```

または明示的にパッケージをプロジェクトに`go get`ください。

    go get github.com/payjp/payjp-go@beta

## Documentation

Please see our official [documentation](http://pay.jp/docs/api/?go).
