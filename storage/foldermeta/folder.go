// Package foldermeta предоставляет значки и фоновые ресурсы папок для приложений Granio.
package foldermeta

import (
	"github.com/Uzer007-admin/granio/v2"
	metadata "github.com/Uzer007-admin/granio/v2/internal/foldermeta"
)

// FancyFolder содержит проверенные метаданные представления папки.
type FancyFolder = metadata.FancyFolder

var ErrNoMetadata = metadata.ErrNoMetadata

// DetailsForFolder возвращает стандартный значок либо доступный .background.*.
func DetailsForFolder(uri granio.URI) (*FancyFolder, error) {
	return metadata.DetailsForFolder(uri)
}
