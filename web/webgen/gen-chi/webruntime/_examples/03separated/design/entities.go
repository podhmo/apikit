package design

type Article struct {
	ID       int64      `json:"id"`
	Title    string     `json:"title"`
	Text     string     `json"text"`
	Comments []*Comment `json:"comments"`
}

type Comment struct {
	Author string `json:"author"`
	Text   string `json"text"`
}
