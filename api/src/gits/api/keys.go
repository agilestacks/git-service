package api

func UsersBySshKey(keyBase64 string, keyFingerprintSHA256 string) ([]string, error) {
	return []string{"anton", "arkadi", "igor", "igorlysak", "nikolay", "oleg", "rick"}, nil
}
