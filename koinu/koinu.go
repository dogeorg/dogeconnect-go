package koinu

import (
	"fmt"
	"strings"
)

// Koinu is the smallest unit of Dogecoin.
// This type is used to represent currency in transactions.
type Koinu int64

// OneDoge is one Dogecoin represented in Koinu, the smallest unit of Dogecoin.
const OneDoge = 100_000_000 // 100 million Koinu (8 zeroes)

// MaxMoney is the maximum Koinu value that can be used in a transaction.
const MaxMoney = 10_000_000_000 * OneDoge // max transaction is 10,000,000,000 Doge

// String implements fmt.Stringer
func (val Koinu) String() string {
	if val < 0 {
		// Negate the quotient and remainder separately to avoid overflow.
		// -val overflows for math.MinInt64, but -(val/OneDoge) and -(val%OneDoge)
		// are both small enough to negate safely.
		whole := -(val / OneDoge)
		part := -(val % OneDoge)
		if part != 0 {
			return strings.TrimRight(fmt.Sprintf("-%d.%08d", whole, part), "0")
		}
		return fmt.Sprintf("-%d", whole)
	}
	whole := val / OneDoge
	part := val % OneDoge
	if part != 0 {
		return strings.TrimRight(fmt.Sprintf("%d.%08d", whole, part), "0")
	}
	return fmt.Sprintf("%d", whole)
}
