package api

type Config struct {
	Addr          string
	Token         string
	Root          string
	FrontUser     string
	FrontPassword string
}

type Page struct {
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	Type     string         `json:"type"`
	Stage    string         `json:"stage"`
	Sorting  string         `json:"sorting"`
	Content  []string       `json:"content"`
	Metadata map[string]any `json:"metadata"`
	Children []*Page        `json:"children,omitempty"`
}
