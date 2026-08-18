package global

const (
	// IssuerYYGU yygu
	IssuerYYGU = "yygu"
	// IssuerYYGUAdmin yygu-admin
	IssuerYYGUAdmin = "yyguadmin"
	// IssuerH3C h3c
	IssuerH3C = "h3c"
)

// IssuerPublicKeyMap is map issuer to license's rsa public key
// this value shouldn't be modified at any time
// unless append a new key
var IssuerPublicKeyMap = map[string]string{
	IssuerYYGU: `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA2yzBf3x0puJOadJpnvyqNR0RCRMC5TSUO1EnY5SjDjtqHbap2Lkn
96kkvcyT2sEY/Wr4/KwhkBRET5uZnvzEOQYiyQJIJbVtVQ+jjuUMiOih2PppWt34
9ypLjhcbzAwanbUMM2kfsI8yPBem0fm6ewgKhJVOQWqG8p13+Jnr39a42vKh9M3Q
IndbNOSEvdX77L+73n98UlIpgrkzj0383jHCXy3hTL61ZOjk8rqhGjTorq/LaQQt
Wyc1h8nyKDAnm6RX40BzzP+TiaZyW6sgmK0ugUcMSQMy4S542DgG/VnAayYLMCl3
WMfu/sTo2IfTqfZrTDCpA9+5HcAb4ordHwIDAQAB
-----END RSA PUBLIC KEY-----`,
	IssuerH3C: `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAs4TJy9Cora3DgsEjcDBpoSoqOlJdLPMFYDmujG3XLPVjoLWTlg36
TwX2SgTdWCm97zqSAi1scZssb14X2XktkibVVKTBc6OCFCxQTYEqQon5ImZ2f3c+
NiMP2Ji71SyHFnGd9lR7YqR+JrvnNLD+jdLZvfYDHVJSzpG54RPJA/l8/he61ynu
egyM/WonICtJ65S7w7gqHzoRNVzRAxlogTmKPoGNUN7yZXSO7E2x2Nl6Tvx6S3EU
m/vaekwufoi7gTX7qObHTUjYan44VpMDLfdx+dHfBvv7SFnKL0Tk+zBMNMa/CksU
9BW8UCuzhVuYABZBuhp2J/um2B0WHSFVQwIDAQAB
-----END RSA PUBLIC KEY-----`,
}

// LicenseVersionKeyMap
type (
	Module string
)

var VersionKeyMap = map[string][]Module{
	// all version
	"all": {ModuleChat, ModuleForest, ModuleAgent, ModuleGraph},
	// agent version
	"agent": {ModuleChat, ModuleForest, ModuleAgent},
}

var (
	ModuleChat   Module = "chat"
	ModuleForest Module = "forest"
	ModuleAgent  Module = "agent"
	ModuleGraph  Module = "graph"
	ModuleWrite  Module = "write"
)
