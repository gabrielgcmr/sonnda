package identity

type Identity struct {
	Provider string // ex: "firebase"
	Subject  string // firebase uid
	Email    string
	Claims   map[string]any
}
