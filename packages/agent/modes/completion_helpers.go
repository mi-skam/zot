package modes

func (i *Interactive) completionContext() completionContext {
	return completionContext{Input: i.ed.Value(), CWD: i.cfg.CWD}
}

func (i *Interactive) completionActive() bool {
	if i.completion == nil {
		return false
	}
	return i.completion.Refresh(i.completionContext())
}
