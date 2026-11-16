package web

import "simpleca/internal/ca"

var (
	ConfigC       string = "FR"
	ConfigST      string = "France"
	ConfigL       string = "Paris"
	ConfigO       string = "MyOrg"
	ConfigOU      string = "MyUnit"
	ConfigSize    int    = 2048
	ConfigDays    int    = ca.MaxCertDays
	ConfigKeyType string = "rsa"
)
