package match

type Rules struct {
	IgnoreQuery bool
	IgnoreHost  bool
	StripAuth   bool
}

func DefaultRules() Rules {
	return Rules{IgnoreQuery: false, IgnoreHost: false, StripAuth: true}
}
