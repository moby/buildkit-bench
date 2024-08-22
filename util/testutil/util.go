package testutil

type MultiCloser struct {
	fns []func() error
}

func (mc *MultiCloser) F() func() error {
	return func() error {
		var err error
		for i := range mc.fns {
			if err1 := mc.fns[len(mc.fns)-1-i](); err == nil {
				err = err1
			}
		}
		mc.fns = nil
		return err
	}
}

func (mc *MultiCloser) Append(f func() error) {
	mc.fns = append(mc.fns, f)
}
