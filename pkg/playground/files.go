package playground

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Example struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Files       []File `json:"files,omitempty"`
}

// Files map is modified by ./generated.go created during ./hack/build.sh
var Files = map[string]File{}
var Examples = []Example{}
