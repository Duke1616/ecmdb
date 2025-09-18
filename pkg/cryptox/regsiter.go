package cryptox

type CryptoRegistry struct {
	User     *CryptoManager[string]
	Resource *CryptoManager[string]
	Runner   *CryptoManager[string]
}
