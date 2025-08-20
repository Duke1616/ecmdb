package ldapx

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/ecodeclub/ekit/slice"
	"github.com/go-ldap/ldap/v3"
)

const specialLDAPRunes = ",#+<>;\"="

type LdapProvider interface {
	CheckConnect() error
	VerifyUserCredentials(username string, password string) (domain.Profile, error)
	FindUserDetail(username string) (domain.Profile, error)
	SearchUserWithPaging() ([]domain.Profile, error)
}

type ldapX struct {
	conf Config
}

func (p *ldapX) SearchUserWithPaging() ([]domain.Profile, error) {
	adminClient, err := p.connect(p.conf.BindDN, p.conf.BindPassword)
	if err != nil {
		return nil, err
	}
	defer adminClient.Close()

	searchRequest := ldap.NewSearchRequest(
		p.conf.BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, p.conf.SyncUserFilter, []string{"dn",
			p.conf.MailAttribute,
			p.conf.UsernameAttribute,
			p.conf.DisplayNameAttribute,
			p.conf.TitleAttribute,
			p.conf.WhenCreatedAttribute,
		}, nil,
	)

	sr, err := adminClient.SearchWithPaging(searchRequest, 10)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	userProfiles := slice.FilterMap(sr.Entries, func(idx int, src *ldap.Entry) (domain.Profile, bool) {
		// 排除指定 ou 的数据
		if strings.Contains(src.DN, p.conf.SyncExcludeOu) {
			return domain.Profile{}, false
		}

		userProfile := domain.Profile{}
		userProfile.DN = src.DN
		for _, attr := range src.Attributes {
			if attr.Name == p.conf.MailAttribute {
				userProfile.Email = attr.Values[0]
			}

			if attr.Name == p.conf.UsernameAttribute {
				if len(attr.Values) != 1 {
					return domain.Profile{}, false
				}

				userProfile.Username = attr.Values[0]
			}
			if attr.Name == p.conf.DisplayNameAttribute {
				userProfile.DisplayName = attr.Values[0]
			}

			if attr.Name == p.conf.TitleAttribute {
				userProfile.Title = attr.Values[0]
			}

			if attr.Name == p.conf.WhenCreatedAttribute {
				userProfile.WhenCreated = attr.Values[0]
			}
		}

		return userProfile, true
	})

	return userProfiles, nil
}

func NewLdap(conf Config) LdapProvider {
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

func (p *ldapX) VerifyUserCredentials(username string, password string) (domain.Profile, error) {
	adminClient, err := p.connect(p.conf.BindDN, p.conf.BindPassword)
	if err != nil {
		return domain.Profile{}, err
	}
	defer adminClient.Close()

	profile, err := p.getUserProfile(adminClient, username)
	if err != nil {
		return domain.Profile{}, err
	}

	conn, err := p.connect(profile.DN, password)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("authentication of user %s failed. Cause: %s", password, err)
	}
	defer conn.Close()

	return profile, nil
}

func (p *ldapX) getUserProfile(conn Connection, inputUsername string) (domain.Profile, error) {
	userFilter := p.resolveUserFilter(p.conf.UserFilter, inputUsername)
	slog.Debug("Computed user filter is %s", userFilter)

	baseDN := p.conf.BaseDN

	attributes := []string{"dn",
		p.conf.MailAttribute,
		p.conf.UsernameAttribute,
		p.conf.DisplayNameAttribute,
		p.conf.TitleAttribute,
	}

	// Search for the given username.
	searchRequest := ldap.NewSearchRequest(
		baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, userFilter, attributes, nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("cannot find user DN of user %s. Cause: %s", inputUsername, err)
	}

	if len(sr.Entries) == 0 {
		return domain.Profile{}, fmt.Errorf("user not found")
	}

	if len(sr.Entries) > 1 {
		return domain.Profile{}, fmt.Errorf("multiple users %s found", inputUsername)
	}

	userProfile := domain.Profile{
		DN: sr.Entries[0].DN,
	}

	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == p.conf.MailAttribute {
			userProfile.Email = attr.Values[0]
		}

		if attr.Name == p.conf.UsernameAttribute {
			if len(attr.Values) != 1 {
				return domain.Profile{}, fmt.Errorf("user %s cannot have multiple value for attribute %s",
					inputUsername, p.conf.UsernameAttribute)
			}

			userProfile.Username = attr.Values[0]
		}
		if attr.Name == p.conf.DisplayNameAttribute {
			userProfile.DisplayName = attr.Values[0]
		}

		if attr.Name == p.conf.TitleAttribute {
			userProfile.Title = attr.Values[0]
		}
	}

	if userProfile.DN == "" {
		return domain.Profile{}, fmt.Errorf("no DN has been found for user %s", inputUsername)
	}

	return userProfile, nil
}

func (p *ldapX) resolveUserFilter(userFilter string, username string) string {
	username = p.ldapEscape(username)

	// We temporarily keep placeholder {0} for backward compatibility.
	userFilter = strings.ReplaceAll(userFilter, "{0}", username)

	// The {username} placeholder is equivalent to {0}, it's the new way, a named placeholder.
	userFilter = strings.ReplaceAll(userFilter, "{input}", username)

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

func (p *ldapX) FindUserDetail(username string) (domain.Profile, error) {
	conn, err := p.connect(p.conf.BindDN, p.conf.BindPassword)
	if err != nil {
		return domain.Profile{}, err
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, username)
	if err != nil {
		return domain.Profile{}, err
	}

	GroupFilter, err := p.resolveGroupFilter(p.conf.GroupFilter, username, &profile)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("unable to create group filter for user %s. Cause: %s", username, err)
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
		return domain.Profile{}, fmt.Errorf("unable to retrieve groups of user %s. Cause: %s", username, err)
	}

	for _, res := range sr.Entries {
		if len(res.Attributes) == 0 {
			slog.Debug("No groups retrieved from LDAP for user %s", username)
			break
		}
		// Append all values of the document. Normally there should be only one per document.
		profile.Groups = append(profile.Groups, res.Attributes[0].Values...)
	}

	return profile, nil
}

func (p *ldapX) resolveGroupFilter(groupFilter, username string, profile *domain.Profile) (string, error) { //nolint:unparam
	username = p.ldapEscape(username)

	// We temporarily keep placeholder {0} for backward compatibility.
	groupFilter = strings.ReplaceAll(groupFilter, "{0}", username)
	groupFilter = strings.ReplaceAll(groupFilter, "{input}", username)

	if profile != nil {
		// We temporarily keep placeholder {1} for backward compatibility.
		groupFilter = strings.ReplaceAll(groupFilter, "{1}", ldap.EscapeFilter(profile.Username))
		groupFilter = strings.ReplaceAll(groupFilter, "{username}", ldap.EscapeFilter(profile.Username))
		groupFilter = strings.ReplaceAll(groupFilter, "{dn}", ldap.EscapeFilter(profile.DN))
	}

	return groupFilter, nil
}
