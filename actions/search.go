package actions

import (
	"../config"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v2"
	"strings"
)

func GetLdapValues(uid string, attributes []string) ([]string, error) {
	uidSplit := strings.Split(uid, "@")
	// config.C.LdapUrl looks like fqdn:portNo
	l, err := ldap.Dial("tcp", config.C.LdapUrl)
	if err != nil {
		log.WithFields(log.Fields{
			"uid":        uidSplit[0],
			"attrbiutes": attributes,
			"LdapUrl":    config.C.LdapUrl,
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
			"LdapUrl":    config.C.LdapUrl,
			"state":      "searching",
		}).Warn(err)
	}
	// return if error encountered here so application can continue.
	attrResults := make([]string, len(attributes))
	for _, entry := range searchResults.Entries {
		for i, j := range attributes {
			attrResults[i] = entry.GetAttributeValue(j)
		}
	}
	return attrResults, err
}
