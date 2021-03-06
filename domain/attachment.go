package domain

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/purwandi/mail"
	"github.com/segmentio/ksuid"
)

// Attachment ...
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Filepath string `json:"filepath"`
	Type     string `json:"type"`
}

// CreateAttachment for storing and creating attachment data
func CreateAttachment(filename, contentType string, content []byte) (Attachment, error) {
	fid := ksuid.New().String()
	attch := Attachment{
		ID:       fid,
		Filename: filename,
		Filepath: fmt.Sprintf("%s%s", fid, filepath.Ext(filename)),
		Type:     contentType,
	}

	// write into file
	f, err := os.Create(fmt.Sprintf("%s/%s", mail.AssetFilePath, attch.Filepath))
	if err != nil {
		return Attachment{}, err
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		return Attachment{}, err
	}

	return attch, nil
}
