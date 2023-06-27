package longpoll

import "os"

type File struct {
	FileInfo    os.FileInfo
	AbsolutPath string
	Name        string
}

func NewFile(path string) (*File, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &File{
		FileInfo:    fileInfo,
		AbsolutPath: path,
		Name:        fileInfo.Name(),
	}, nil
}
