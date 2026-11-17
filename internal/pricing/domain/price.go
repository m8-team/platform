package domain

import (
	"errors"
	"math/big"
)

var (
	ErrMissingPrice       = errors.New("price and discount are required")
	ErrNegativePrice      = errors.New("price must not be negative")
	ErrDiscountOutOfRange = errors.New("discount percentage must be between 0 and 100")
)

// CalculateSellingPrice returns price * (1 - discount/100) exactly, without
// currency rounding. Both arguments are decimal values (not binary floats),
// and discount is expressed in percent: 80.01 means an 80.01% discount.
func CalculateSellingPrice(price, discount *big.Rat) (*big.Rat, error) {
	if price == nil || discount == nil {
		return nil, ErrMissingPrice
	}
	if price.Sign() < 0 {
		return nil, ErrNegativePrice
	}
	hundred := big.NewRat(100, 1)
	if discount.Sign() < 0 || discount.Cmp(hundred) > 0 {
		return nil, ErrDiscountOutOfRange
	}
	factor := new(big.Rat).Sub(big.NewRat(1, 1), new(big.Rat).Quo(discount, hundred))
	return new(big.Rat).Mul(price, factor), nil
}
