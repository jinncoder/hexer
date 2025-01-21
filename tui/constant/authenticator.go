package constant

type AuthenticatorType int

var AuthenticatorTypeList = [...]string{"Host", "Domain", "Application", "Other"}

const (
	AuthenticatorTypeHost AuthenticatorType = iota
	AuthenticatorTypeDomain
	AuthenticatorTypeApplication
	AuthenticatorTypeOther
)

func (c AuthenticatorType) String() string {
	return AuthenticatorTypeList[c]
}
