package fail

type OnErrors struct {
	Max        int
	errCounter int

	recent error
}

func (o *OnErrors) Record(err error) {
	if err != nil {
		o.errCounter++
		o.recent = err
		return
	}

	o.errCounter = max(0, o.errCounter-1)
}

func (o *OnErrors) Fail() bool {
	return o.errCounter >= o.Max
}

func (o *OnErrors) Recent() error {
	return o.recent
}
