package organization

type Item struct {
	Login       string
	Description string
	URL         string
}

func (i Item) FilterValue() string {
	return i.Login
}
