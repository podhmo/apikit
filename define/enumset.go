package define

type EnumSet struct {
	Name  string
	Enums []Enum
}

type Enum struct {
	Name        string
	Value       interface{}
	Description string
}
