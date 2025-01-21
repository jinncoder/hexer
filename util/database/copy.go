package database

import (
	"fmt"
	"strings"

	"github.com/archimoebius/hexer/tui/constant"
)

func CredentialFormImpacket(authenticator Authenticator, credential Credential) []string {
	var credentialForms []string

	var FQDNOrIPAddress = authenticator.FQDN
	if FQDNOrIPAddress == "" {
		FQDNOrIPAddress = authenticator.IPv4
	}

	switch strings.ToLower(credential.Type) {
	case strings.ToLower(constant.SCredentialTypeCleartext):
		credentialForms = append(credentialForms, fmt.Sprintf("impacket '%s/%s:%s@%s'", authenticator.DomainName(), credential.UserName, credential.Value, FQDNOrIPAddress))
	case strings.ToLower(constant.SCredentialTypeHash):
		if strings.ToLower(credential.Format) == constant.SCredentialFormatNTLM {
			credentialForms = append(credentialForms, fmt.Sprintf("impacket -hashes :%s '%s/%s@%s'", credential.Value, authenticator.DomainName(), credential.UserName, FQDNOrIPAddress))
		}
	case strings.ToLower(constant.SCredentialTypeCipher):
		if strings.ToLower(credential.Format) == constant.SCredentialFormatAES128 || credential.Format == constant.SCredentialFormatAES256 {
			credentialForms = append(credentialForms, fmt.Sprintf("impacket -aesKey %s '%s/%s@%s'", credential.Value, authenticator.DomainName(), credential.UserName, FQDNOrIPAddress))
		}
	}

	return credentialForms
}

func CredentialFormNetExec(authenticator Authenticator, credential Credential) []string {
	var credentialForms []string

	var FQDNOrIPAddress = authenticator.FQDN
	if FQDNOrIPAddress == "" {
		FQDNOrIPAddress = authenticator.IPv4
	}

	switch strings.ToLower(credential.Type) {
	case strings.ToLower(constant.SCredentialTypeCleartext):
		credentialForms = append(credentialForms, fmt.Sprintf("netexec %s --domain '%s' --username '%s'  --password '%s'", FQDNOrIPAddress, authenticator.DomainName(), credential.UserName, credential.Value))
	case strings.ToLower(constant.SCredentialTypeHash):
		if strings.ToLower(credential.Format) == constant.SCredentialFormatNTLM {
			credentialForms = append(credentialForms, fmt.Sprintf("netexec %s -H :%s --domain '%s' --username '%s'", FQDNOrIPAddress, credential.Value, authenticator.DomainName(), credential.UserName))
		}
	case strings.ToLower(constant.SCredentialTypeCipher):
		if strings.ToLower(credential.Format) == constant.SCredentialFormatAES128 || credential.Format == constant.SCredentialFormatAES256 {
			credentialForms = append(credentialForms, fmt.Sprintf("netexec %s --aesKey %s --domain '%s' --username '%s'", FQDNOrIPAddress, credential.Value, authenticator.DomainName(), credential.UserName))
		}
	}

	return credentialForms
}

func CredentialFormEvilWinRM(authenticator Authenticator, credential Credential) []string {
	var credentialForms []string

	var FQDNOrIPAddress = authenticator.FQDN
	if FQDNOrIPAddress == "" {
		FQDNOrIPAddress = authenticator.IPv4
	}

	switch strings.ToLower(credential.Type) {
	case strings.ToLower(constant.SCredentialTypeCleartext):
		credentialForms = append(credentialForms, fmt.Sprintf("evil-winrm --ip %s --realm '%s' --user '%s'  --password '%s'", FQDNOrIPAddress, authenticator.DomainName(), credential.UserName, credential.Value))
	case strings.ToLower(constant.SCredentialTypeHash):
		if strings.ToLower(credential.Format) == constant.SCredentialFormatNTLM {
			credentialForms = append(credentialForms, fmt.Sprintf("evil-winrm --ip %s --hash %s --realm '%s' --user '%s'", FQDNOrIPAddress, credential.Value, authenticator.DomainName(), credential.UserName))
		}
	}

	return credentialForms
}
