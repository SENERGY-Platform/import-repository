package source

import "github.com/SENERGY-Platform/import-repository/lib/model"

type ImportTypeCommand struct {
	Command    string                   `json:"command"`
	Id         string                   `json:"id"`
	Owner      string                   `json:"owner"`
	ImportType model.ImportTypeExtended `json:"import_type"`
}
