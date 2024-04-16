package ldapx

import (
	"crypto/tls"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"log/slog"
	"net/url"
	"strings"
)

const specialLDAPRunes = ",#+<>;\"="

type LdapInterface interface {
	CheckConnect() error
	CheckUserPassword(username string, password string) (*Profile, error)
	GetDetails(username string) (*Profile, error)
}

type ldapX struct {
	conf Config
}

func NewLdap(conf Config) LdapInterface {
	return &ldapX{
		conf: conf,
	}
}

func (p *ldapX) CheckConnect() error {
	adminClient, err := p.connect(p.conf.BindDN, p.conf.BindPassword)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	return nil
}

func (p *ldapX) connect(userDN string, password string) (Connection, error) {
	var conn Connection

	urlPath, err := url.Parse(p.conf.Url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL to LDAP: %s", urlPath)
	}

	if urlPath.Scheme == "ldaps" {
		slog.Debug("LDAP client starts a TLS session")
		tlsConn, err := p.dialTLS(p.conf.Url, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return nil, err
		}

		conn = tlsConn
	} else {
		slog.Debug("LDAP client starts a session over raw TCP")
		rawConn, err := p.dial(p.conf.Url)
		if err != nil {
			return nil, err
		}
		conn = rawConn
	}

	if err = conn.Bind(userDN, password); err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *ldapX) dialTLS(addr string, config *tls.Config) (Connection, error) {
	conn, err := ldap.DialURL(addr, ldap.DialWithTLSConfig(config))
	if err != nil {
		return nil, err
	}

	return NewLDAPConnectionImpl(conn), nil
}

func (p *ldapX) dial(addr string) (Connection, error) {
	conn, err := ldap.DialURL(addr)
	if err != nil {
		return nil, err
	}

	return NewLDAPConnectionImpl(conn), nil
}

func (p *ldapX) CheckUserPassword(username string, password string) (*Profile, error) {
	adminClient, err := p.connect(p.conf.BindDN, p.conf.BindPassword)
	if err != nil {
		return nil, err
	}
	defer adminClient.Close()

	profile, err := p.getUserProfile(adminClient, username)
	if err != nil {
		return nil, err
	}

	conn, err := p.connect(profile.DN, password)
	if err != nil {
		return nil, fmt.Errorf("authentication of user %s failed. Cause: %s", password, err)
	}
	defer conn.Close()

	return profile, nil
}

func (p *ldapX) getUserProfile(conn Connection, inputUsername string) (*Profile, error) {
	userFilter := p.resolveUserFilter(p.conf.UserFilter, inputUsername)
	slog.Debug("Computed user filter is %s", userFilter)

	baseDN := p.conf.BaseDN

	attributes := []string{"dn",
		p.conf.MailAttribute,
		p.conf.UsernameAttribute,
		p.conf.DisplayNameAttribute,
	}

	// Search for the given username.
	searchRequest := ldap.NewSearchRequest(
		baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, userFilter, attributes, nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot find user DN of user %s. Cause: %s", inputUsername, err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("multiple users %s found", inputUsername)
	}

	userProfile := Profile{
		DN: sr.Entries[0].DN,
	}

	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == p.conf.MailAttribute {
			userProfile.Email = attr.Values[0]
		}

		if attr.Name == p.conf.UsernameAttribute {
			if len(attr.Values) != 1 {
				return nil, fmt.Errorf("user %s cannot have multiple value for attribute %s",
					inputUsername, p.conf.UsernameAttribute)
			}

			userProfile.Username = attr.Values[0]
		}
		if attr.Name == p.conf.DisplayNameAttribute {
			userProfile.DisplayName = attr.Values[0]
		}
	}

	if userProfile.DN == "" {
		return nil, fmt.Errorf("no DN has been found for user %s", inputUsername)
	}

	return &userProfile, nil
}

func (p *ldapX) resolveUserFilter(userFilter string, inputUsername string) string {
	inputUsername = p.ldapEscape(inputUsername)

	// We temporarily keep placeholder {0} for backward compatibility.
	userFilter = strings.ReplaceAll(userFilter, "{0}", inputUsername)

	// The {username} placeholder is equivalent to {0}, it's the new way, a named placeholder.
	userFilter = strings.ReplaceAll(userFilter, "{input}", inputUsername)

	// {username_attribute} and {mail_attribute} are replaced by the content of the attribute defined
	// in configuration.
	userFilter = strings.ReplaceAll(userFilter, "{username_attribute}", p.conf.UsernameAttribute)
	userFilter = strings.ReplaceAll(userFilter, "{mail_attribute}", p.conf.MailAttribute)
	return userFilter
}

func (p *ldapX) ldapEscape(inputUsername string) string {
	inputUsername = ldap.EscapeFilter(inputUsername)
	for _, c := range specialLDAPRunes {
		inputUsername = strings.ReplaceAll(inputUsername, string(c), fmt.Sprintf("\\%c", c))
	}

	return inputUsername
}

func (p *ldapX) GetDetails(inputUsername string) (*Profile, error) {
	conn, err := p.connect(p.conf.BindDN, p.conf.BindPassword)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)
	if err != nil {
		return nil, err
	}

	GroupFilter, err := p.resolveGroupFilter(inputUsername, profile)
	if err != nil {
		return nil, fmt.Errorf("unable to create group filter for user %s. Cause: %s", inputUsername, err)
	}

	slog.Debug("Computed groups filter is %s", GroupFilter)

	groupBaseDN := p.conf.BaseDN

	// Search for the given username.
	searchGroupRequest := ldap.NewSearchRequest(
		groupBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, GroupFilter, []string{p.conf.GroupNameAttribute}, nil,
	)

	sr, err := conn.Search(searchGroupRequest)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve groups of user %s. Cause: %s", inputUsername, err)
	}

	for _, res := range sr.Entries {
		if len(res.Attributes) == 0 {
			slog.Debug("No groups retrieved from LDAP for user %s", inputUsername)
			break
		}
		// Append all values of the document. Normally there should be only one per document.
		profile.Groups = append(profile.Groups, res.Attributes[0].Values...)
	}

	return profile, nil
}

func (p *ldapX) resolveGroupFilter(inputUsername string, profile *Profile) (string, error) { //nolint:unparam
	inputUsername = p.ldapEscape(inputUsername)

	// We temporarily keep placeholder {0} for backward compatibility.
	groupFilter := strings.ReplaceAll(p.conf.GroupFilter, "{0}", inputUsername)
	groupFilter = strings.ReplaceAll(groupFilter, "{input}", inputUsername)

	if profile != nil {
		// We temporarily keep placeholder {1} for backward compatibility.
		groupFilter = strings.ReplaceAll(groupFilter, "{1}", ldap.EscapeFilter(profile.Username))
		groupFilter = strings.ReplaceAll(groupFilter, "{username}", ldap.EscapeFilter(profile.Username))
		groupFilter = strings.ReplaceAll(groupFilter, "{dn}", ldap.EscapeFilter(profile.DN))
	}

	return groupFilter, nil
}
