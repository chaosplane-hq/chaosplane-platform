package service

func SignTestAccessToken(claims *AccessTokenClaims, secret string) (string, error) {
	return signToken(claims, secret)
}
