package generate

type Generator struct {
	RootDirs    map[string]string
	RootPackage string
}

func NewGenerator(rootdir string) *Generator {
	return &Generator{
		RootDirs: map[string]string{"": rootdir},
	}
}
