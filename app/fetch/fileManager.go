package fetch

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type FileManager struct {
	HTML    string
	OldHtml string
	links   []ResourceLink
	Table   map[string]string
}

func NewFileManager(html string, links []ResourceLink) *FileManager {
	return &FileManager{
		HTML:    html,
		OldHtml: html,
		links:   links,
		Table:   make(map[string]string),
	}
}

func (fm *FileManager) ReplaceLinks() {
	for _, link := range fm.links {
		uuid := uuid.New().String()
		fm.Table[link.Absolute] = uuid
		fm.HTML = strings.ReplaceAll(fm.HTML, link.Original, fmt.Sprintf("./contents/%s", uuid))
	}
}
