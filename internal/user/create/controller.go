package create

type inputData struct {
	FirstName string `json:"first_name" validate:"required,min=3,max=100"`
	LastName  string `json:"last_name" validate:"required,min=3,max=100"`
}
