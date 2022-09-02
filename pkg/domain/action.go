package domain

type Action struct {
	Name string
	Code string
}

var (
	DocumentPickup = Action{Name: "Document pickup", Code: "DOC"}
	Biometrics     = Action{Name: "Biometrics", Code: "BIO"}
)
