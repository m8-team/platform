package domain

import (
	"errors"
	"math/big"
	"testing"
)

func mustDecimal(t *testing.T, value string) *big.Rat {
	t.Helper()
	v, ok := new(big.Rat).SetString(value)
	if !ok {
		t.Fatalf("invalid decimal %q", value)
	}
	return v
}

func TestCalculateSellingPrice(t *testing.T) {
	for _, tc := range []struct {
		name     string
		price    string
		discount string
		want     string
	}{
		{name: "zero discount", price: "9999.00", discount: "0", want: "9999.00"},
		{name: "retail markdown", price: "9999.00", discount: "80.01", want: "1998.8001"},
		{name: "other markdown", price: "3499.00", discount: "85.74", want: "498.9574"},
		{name: "full discount", price: "9999.00", discount: "100", want: "0"},
		{name: "zero rrp", price: "0", discount: "25", want: "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			price := mustDecimal(t, tc.price)
			discount := mustDecimal(t, tc.discount)
			oldPrice := new(big.Rat).Set(price)
			oldDiscount := new(big.Rat).Set(discount)
			got, err := CalculateSellingPrice(price, discount)
			if err != nil {
				t.Fatalf("CalculateSellingPrice() error: %v", err)
			}
			if got.Cmp(mustDecimal(t, tc.want)) != 0 {
				t.Errorf("selling price = %v, want %s", got, tc.want)
			}
			if price.Cmp(oldPrice) != 0 || discount.Cmp(oldDiscount) != 0 {
				t.Error("input arguments mutated")
			}
		})
	}
}

func TestCalculateSellingPriceErrors(t *testing.T) {
	cases := []struct {
		name     string
		price    *big.Rat
		discount *big.Rat
		want     error
	}{
		{name: "missing price", discount: big.NewRat(5, 1), want: ErrMissingPrice},
		{name: "missing discount", price: big.NewRat(10, 1), want: ErrMissingPrice},
		{name: "negative price", price: big.NewRat(-1, 1), discount: big.NewRat(5, 1), want: ErrNegativePrice},
		{name: "negative discount", price: big.NewRat(10, 1), discount: big.NewRat(-1, 1), want: ErrDiscountOutOfRange},
		{name: "over 100 percent", price: big.NewRat(10, 1), discount: big.NewRat(101, 1), want: ErrDiscountOutOfRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CalculateSellingPrice(tc.price, tc.discount)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error %v, want %v", err, tc.want)
			}
		})
	}
}
