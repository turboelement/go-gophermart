package utils

import "strconv"

func IsValidLuhn(number string) bool {
	nDigits := len(number)
	if nDigits < 2 {
		return false
	}

	sum := 0
	for i, n := range number {
		digit, err := strconv.Atoi(string(n))
		if err != nil {
			return false // not number
		}

		if (nDigits-1-i)%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}
