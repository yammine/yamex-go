package yamex

type Sentinel string

func (s Sentinel) Error() string { return string(s) }
