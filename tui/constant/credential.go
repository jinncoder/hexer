package constant

// CredentialType -----------------------------------------------------------------------

// Define the CredentialType type
type CredentialType int

// List of credential types as strings
var CredentialTypeList = [...]string{"Cleartext", "Hash", "Cipher", "Other", "Null"}

// Define the CredentialType constants
const (
	CredentialTypeCleartext CredentialType = iota
	CredentialTypeHash
	CredentialTypeCipher
	CredentialTypeOther
	CredentialTypeNull
)

const (
	SCredentialTypeCleartext = "Cleartext"
	SCredentialTypeHash      = "Hash"
	SCredentialTypeCipher    = "Cipher"
	SCredentialTypeOther     = "Other"
	SCredentialTypeNull      = "Null"
)

var SCredentialTypeMap = initSCredentialTypeMap()

func initSCredentialTypeMap() map[string]CredentialType {
	return map[string]CredentialType{
		SCredentialTypeCleartext: CredentialTypeCleartext,
		SCredentialTypeHash:      CredentialTypeHash,
		SCredentialTypeCipher:    CredentialTypeCipher,
		SCredentialTypeOther:     CredentialTypeOther,
		SCredentialTypeNull:      CredentialTypeNull,
	}

}
func (c CredentialType) String() string {
	return CredentialTypeList[c]
}

// CredentialFormat ---------------------------------------------------------------------

type CredentialFormat int

var CredentialFormatList = [...]string{"Raw", "NTLM", "SSH-RSA", "SSH-DSA", "SSH-ED", "AES-128", "AES-256", "Other"}

const (
	CredentialFormatRaw CredentialFormat = iota
	CredentialFormatNTLM
	CredentialFormatSSH_RSA
	CredentialFormatSSH_DSA
	CredentialFormatSSH_ED
	CredentialFormatAES128
	CredentialFormatAES256
	CredentialFormatRC4
	CredentialFormatDES3
	CredentialFormatOther
)

const (
	SCredentialFormatRaw     = "Raw"
	SCredentialFormatNTLM    = "NTLM"
	SCredentialFormatSSH_RSA = "SSH-RSA"
	SCredentialFormatSSH_DSA = "SSH-DSA"
	SCredentialFormatSSH_ED  = "SSH-ED"
	SCredentialFormatAES128  = "AES-128"
	SCredentialFormatAES256  = "AES-256"
	SCredentialFormatOther   = "Other"
)

var CredentialFormatMap = initCredentialFormatMap()

func initCredentialFormatMap() map[string]CredentialFormat {
	return map[string]CredentialFormat{
		SCredentialFormatRaw:     CredentialFormatRaw,
		SCredentialFormatNTLM:    CredentialFormatNTLM,
		SCredentialFormatSSH_RSA: CredentialFormatSSH_RSA,
		SCredentialFormatSSH_DSA: CredentialFormatSSH_DSA,
		SCredentialFormatSSH_ED:  CredentialFormatSSH_ED,
		SCredentialFormatAES128:  CredentialFormatAES128,
		SCredentialFormatAES256:  CredentialFormatAES256,
		SCredentialFormatOther:   CredentialFormatOther,
	}
}

func (c CredentialFormat) String() string {
	return CredentialFormatList[c]
}

// CredentialEncoding -------------------------------------------------------------------

type CredentialEncoding int

var CredentialEncodingList = [...]string{"Raw", "Base64", "Hex", "PEM"}

const (
	CredentialEncodingRaw CredentialEncoding = iota
	CredentialEncodingBase64
	CredentialEncodingHex
	CredentialEncodingPEM
)

const (
	SCredentialEncodingRaw    = "Raw"
	SCredentialEncodingBase64 = "Base64"
	SCredentialEncodingHex    = "Hex"
	SCredentialEncodingPEM    = "PEM"
)

var CredentialEncodingMap = initCredentialEncodingMap()

func initCredentialEncodingMap() map[string]CredentialEncoding {
	return map[string]CredentialEncoding{
		SCredentialEncodingRaw:    CredentialEncodingRaw,
		SCredentialEncodingBase64: CredentialEncodingBase64,
		SCredentialEncodingHex:    CredentialEncodingHex,
		SCredentialEncodingPEM:    CredentialEncodingPEM,
	}
}

func (c CredentialEncoding) String() string {
	return CredentialEncodingList[c]
}
