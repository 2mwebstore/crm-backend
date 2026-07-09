package utils

// USDToKHRRate is a fixed exchange rate (1 USD = 4000 KHR) used across the
// system for converting between USD and KHR when a transaction's currency
// differs from a balance's own currency (e.g. a KHR deposit affecting a
// USD-denominated product credit pool).
//
// NOTE: this is a fixed constant, not a live/market rate. If the business
// needs a rate that changes over time, this should move to a DB-backed
// lookup (e.g. a currency_rates table with an effective date) instead of a
// hardcoded value here.
const USDToKHRRate = 4000.0

// ConvertCurrency converts amount from one currency code to another.
// Only USD <-> KHR is supported; any other combination (equal currencies,
// an empty/unknown code, or any currency pair other than USD/KHR) returns
// amount unchanged rather than guessing at an unsupported conversion.
func ConvertCurrency(amount float64, from, to string) float64 {
	if from == to || from == "" || to == "" {
		return amount
	}
	switch {
	case from == "USD" && to == "KHR":
		return RoundFloat(amount*USDToKHRRate, 2)
	case from == "KHR" && to == "USD":
		return RoundFloat(amount/USDToKHRRate, 2)
	default:
		return amount
	}
}
