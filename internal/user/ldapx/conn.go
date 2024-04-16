package ldapx

import "github.com/go-ldap/ldap/v3"

type Connection interface {
	Bind(username, password string) error
	Close()
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	SearchWithPaging(searchRequest *ldap.SearchRequest, pagingSize uint32) (*ldap.SearchResult, error)
}

type ConnectionImpl struct {
	conn *ldap.Conn
}

func NewLDAPConnectionImpl(conn *ldap.Conn) *ConnectionImpl {
	return &ConnectionImpl{conn}
}

func (lc *ConnectionImpl) Bind(username, password string) error {
	return lc.conn.Bind(username, password)
}

func (lc *ConnectionImpl) Close() {
	lc.conn.Close()
}

func (lc *ConnectionImpl) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	return lc.conn.Search(searchRequest)
}

func (lc *ConnectionImpl) SearchWithPaging(searchRequest *ldap.SearchRequest, pagingSize uint32) (*ldap.SearchResult, error) {
	return lc.conn.SearchWithPaging(searchRequest, pagingSize)
}
