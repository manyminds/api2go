package api2go

//staticResolver is only used
//for backwards compatible reasons
//and might be removed in the future
type staticResolver struct {
	baseURL string
}

func (s staticResolver) GetBaseURL() string {
	return s.baseURL
}

func newStaticResolver(baseURL string) URLResolver {
	return &staticResolver{baseURL: baseURL}
}
