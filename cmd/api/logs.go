package main

func (a *api) notFoundLog(method string, path string, err error) {
	a.logger.Warnw("not found error", "method", method, "path", path, "error", err.Error())
}

func (a *api) badRequestLog(method string, path string, err error) {
	a.logger.Errorf("bad request error", "method", method, "path", path, "error", err.Error())
}

func (a *api) internalErrLog(method string, path string, err error) {
	a.logger.Errorw("internal error", "method", method, "path", path, "error", err.Error())
}

func (a *api) unauthorizedLog(method string, path string, err error) {
	a.logger.Errorf("unauthorized request", "method", method, "path", path, "error", err.Error())
}

func (a *api) conflictLog(method string, path string, err error) {
	a.logger.Errorf("conflicted request", "method", method, "path", path, "error", err.Error())
}

