package longpoll

type State struct {
	Files []File
}

func NewState() *State {
	return &State{}
}

func (s *State) AddFile(file File) {
	s.Files = append(s.Files, file)
}

func (s *State) RemoveFile(file File) {
	for i, f := range s.Files {
		if f.AbsolutPath == file.AbsolutPath {
			s.Files = append(s.Files[:i], s.Files[i+1:]...)
			return
		}
	}
}

func (s *State) GetFile(path string) *File {
	for _, file := range s.Files {
		if file.AbsolutPath == path {
			return &file
		}
	}

	return nil
}
