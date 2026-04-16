package httpclient

// InitHttpClient is the constructor equivalent to InitRestyClient (Wire)
// but designed for manual / Fx initialization.
func InitHttpClient() (HttpClientInterface, error) {
	cfg, err := NewConfig()
	if err != nil {
		return nil, err
	}

	client := NewHttpClient(cfg)
	return client, nil
}
