package koinu

import "errors"

var ErrMaxMoney = errors.New("greater than max-money (10,000,000,000 DOGE)")
var ErrInvalidNumber = errors.New("invalid number (unexpected character)")

const maxMoneyDigits = 11                  // number of integer-part digits in MaxMoney
const maxKoinuDigits = 8                   // number of fraction-part digits in OneDoge
const maxMoneyInteger = MaxMoney / OneDoge // MaxMoney integer-part, to avoid overflow

// decimalScale is 8 powers of 10, reversed for 'length' indexing.
// as a special case, length zero yields a zero.
var decimalScale = []int64{0, 10_000_000, 1_000_000, 100_000, 10_000, 1000, 100, 10, 1}

// ParseKoinu parses a decimal string like "12.6251" to Koinu (1262510000)
// Dogecoin decimals are significant to 8 places at the protocol level
// and are encoded in transactions as 64-bit integers.
func ParseKoinu(amt string) (Koinu, error) {
	chars := []uint8(amt)
	sign := int64(1)
	n := len(amt)
	i := 0

	// optional minus sign
	if i < n && chars[i] == '-' {
		sign = -1
		i++
	}

	// skip leading zeroes (because we limit to maxMoneyDigits)
	start := i
	for i < n && chars[i] == '0' {
		i++
	}

	// whole number part
	whole, i := parseUInt64(chars, i, n, maxMoneyDigits)
	moreDigits := i < n && chars[i]-'0' < 10
	if whole > maxMoneyInteger || moreDigits {
		// whole part is greater than MaxMoney
		return 0, ErrMaxMoney
	}

	digits := i - start // digits consumed so far (including leading zeros)

	whole = whole * OneDoge // overflow: safe due to check above (approx. 1/10 of MaxInt64)

	// decimal part, up to 8 significant digits
	if i < n && chars[i] == '.' {
		decStart := i + 1
		part, end := parseUInt64(chars, decStart, n, maxKoinuDigits)
		i = end

		digits += end - decStart

		// decimal part must be 8 digits; multiply by 10 ^ (8 - length)
		// e.g. if we found 6 digits, multiply by 100
		length := end - decStart
		part *= decimalScale[length]

		whole += part // overflow: safe, less than OneDoge

		// decimal part can push us above MaxMoney
		if whole > MaxMoney {
			return 0, ErrMaxMoney
		}
	}

	// require at least one digit
	if digits == 0 {
		return 0, ErrInvalidNumber
	}

	// invalid if string contains more characters (but skip extra decimal-part digits)
	i = skipDigits(chars, i, n)
	if i != n {
		return 0, ErrInvalidNumber
	}

	return Koinu(sign * whole), nil
}

// parseUInt64 is like strconv.Atoi with int64 result and a maximum length.
func parseUInt64(chars []uint8, i int, n int, maxlen int) (int64, int) {
	n = min(n, i+maxlen)
	val := int64(0)
	for i < n {
		// the following test relies on unsigned modulo math
		ch := chars[i] - '0'
		if ch < 10 {
			val = val*10 + int64(ch)
			i++
		} else {
			break // not a digit
		}
	}
	return val, i
}

func skipDigits(chars []uint8, i int, n int) int {
	for i < n && chars[i]-'0' < 10 {
		i++
	}
	return i
}
