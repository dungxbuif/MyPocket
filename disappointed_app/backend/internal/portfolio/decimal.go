package portfolio

import (
	"math/big"
	"strings"
)

const quantityScale int64 = 1_000_000_000_000

func normalizeDecimal(value string) (string, *big.Rat, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "", nil, ErrValidation
	}
	if strings.HasPrefix(raw, "+") {
		raw = strings.TrimPrefix(raw, "+")
	}
	parts := strings.Split(raw, ".")
	if len(parts) > 2 || parts[0] == "" {
		return "", nil, ErrValidation
	}
	for _, part := range parts {
		if part == "" {
			continue
		}
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return "", nil, ErrValidation
			}
		}
	}
	whole := strings.TrimLeft(parts[0], "0")
	if whole == "" {
		whole = "0"
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
		if frac == "" {
			return "", nil, ErrValidation
		}
		if len(frac) > 12 {
			return "", nil, ErrValidation
		}
		frac = strings.TrimRight(frac, "0")
	}
	normalized := whole
	if frac != "" {
		normalized += "." + frac
	}
	rat := new(big.Rat)
	if _, ok := rat.SetString(normalized); !ok || rat.Sign() <= 0 {
		return "", nil, ErrValidation
	}
	return normalized, rat, nil
}

func parseStoredDecimal(value string) *big.Rat {
	rat := new(big.Rat)
	if _, ok := rat.SetString(strings.TrimSpace(value)); !ok {
		return new(big.Rat)
	}
	return rat
}

func ratToDecimalString(value *big.Rat) string {
	if value == nil || value.Sign() == 0 {
		return "0"
	}
	scaled := new(big.Rat).Mul(value, big.NewRat(quantityScale, 1))
	intScaled := roundRat(scaled)
	whole := new(big.Int).Quo(intScaled, big.NewInt(quantityScale))
	frac := new(big.Int).Mod(intScaled, big.NewInt(quantityScale)).String()
	for len(frac) < 12 {
		frac = "0" + frac
	}
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return whole.String()
	}
	return whole.String() + "." + frac
}

func roundRat(value *big.Rat) *big.Int {
	if value.Sign() == 0 {
		return big.NewInt(0)
	}
	num := new(big.Int).Set(value.Num())
	den := new(big.Int).Set(value.Denom())
	absNum := new(big.Int).Abs(num)
	q, rem := new(big.Int).QuoRem(absNum, den, new(big.Int))
	rem.Mul(rem, big.NewInt(2))
	if rem.Cmp(den) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	if value.Sign() < 0 {
		q.Neg(q)
	}
	return q
}

func roundQuantityMoney(quantity *big.Rat, unitPriceVND int64) (int64, error) {
	if quantity == nil || quantity.Sign() == 0 || unitPriceVND == 0 {
		return 0, nil
	}
	value := new(big.Rat).Mul(quantity, big.NewRat(unitPriceVND, 1))
	return checkedMoney(roundRat(value))
}

func checkedMoney(value *big.Int) (int64, error) {
	if !value.IsInt64() {
		return 0, ErrValidation
	}
	return value.Int64(), nil
}

func addMoney(a, b int64) (int64, error) {
	return checkedMoney(new(big.Int).Add(big.NewInt(a), big.NewInt(b)))
}

func subtractMoney(a, b int64) (int64, error) {
	return checkedMoney(new(big.Int).Sub(big.NewInt(a), big.NewInt(b)))
}

func percentString(numerator int64, denominator int64) *string {
	if denominator == 0 {
		return nil
	}
	value := new(big.Rat).Mul(big.NewRat(numerator, denominator), big.NewRat(100, 1))
	text := value.FloatString(10)
	return &text
}
