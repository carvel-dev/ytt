package workspace

type DataLoader struct {
	library *Library
}

func (l DataLoader) FilePaths() []string {
	result := []string{}
	for _, fileInLib := range l.library.ListAccessibleFiles() {
		result = append(result, fileInLib.RelativePath())
	}
	return result
}

func (l DataLoader) FileData(path string) ([]byte, error) {
	fileInLib, err := l.library.FindFile(path)
	if err != nil {
		return nil, err
	}

	fileBs, err := fileInLib.File.Bytes()
	if err != nil {
		return nil, err
	}

	return fileBs, nil
}
