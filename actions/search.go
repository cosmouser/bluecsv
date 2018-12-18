package actions

import (
	"fmt"
	"strings"

	"github.com/cosmouser/bluecsv/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v2"
)

// GetLdapValues looks up the provided uid and retruns a slice for appending
// in a CSV file
func GetLdapValues(uid string, attributes []string) ([]string, error) {
	uidSplit := strings.Split(uid, "@")
	// config.C.LdapUrl looks like fqdn:portNo
	l, err := ldap.Dial("tcp", config.C.LdapURL)
	if err != nil {
		log.WithFields(log.Fields{
			"uid":        uidSplit[0],
			"attrbiutes": attributes,
			"LdapUrl":    config.C.LdapURL,
			"state":      "dialing",
		}).Warn(err)
	}
	defer l.Close()
	searchReq := ldap.NewSearchRequest(
		config.C.LdapBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(uid=%s))", uidSplit[0]),
		attributes,
		nil,
	)
	searchResults, err := l.Search(searchReq)
	if err != nil {
		log.WithFields(log.Fields{
			"uid":        uidSplit[0],
			"attrbiutes": attributes,
			"LdapUrl":    config.C.LdapURL,
			"state":      "searching",
		}).Warn(err)
	}
	// return if error encountered here so application can continue.
	attrResults := make([]string, len(attributes))
	for _, entry := range searchResults.Entries {
		for i, j := range attributes {
			attrResults[i] = strings.Join(entry.GetAttributeValues(j), "|")
		}
	}
	return attrResults, err
}
