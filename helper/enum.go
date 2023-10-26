package helper

type Sides string

const (
	Yes Sides = "YES"
	No  Sides = "NO"
)

func (s Sides) Valid() string {
	switch s {
	case "YES":
		return string(Yes)
	case "NO":
		return string(No)
	// case "y":
	// 	return string(Yes)
	// case "n":
	// 	return string(No)
	default:
		return "unknown"
	}
}

type FlagOfPurchase string

const (
	Buy  FlagOfPurchase = "BUY"
	Sell FlagOfPurchase = "SELL"
)

func (s FlagOfPurchase) Valid() string {
	switch s {
	case "BUY":
		return string(Buy)
	case "SELL":
		return string(Sell)
	// case "b":
	// 	return string(Buy)
	// case "s":
	// 	return string(Sell)
	default:
		return "unknown"
	}
}
