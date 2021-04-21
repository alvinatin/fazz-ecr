package types

type Cred struct {
	User   string
	Pass   string
	Access []string
	Exp    int64
}
