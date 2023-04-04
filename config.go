package main

type Config struct {
	DebugOn             bool
	Repo                string
	OutputDir           string
	CloneDir            string
	RepoData            RepoData
	AllowedLicenseFiles map[string]bool
	AllowedReadMeFiles  map[string]bool
	TextExtensions      map[string]bool
	PlainFiles          map[string]bool
}

func DefaultConfig() Config {
	return Config{
		DebugOn:   true,
		Repo:      "",
		OutputDir: "",
		CloneDir:  "",
		RepoData: RepoData{
			Name:            "",
			GitURL:          "",
			Description:     "",
			BaseURL:         "/",
			ReadMePath:      "",
			LicenseFilePath: "",
		},
		TextExtensions: map[string]bool{
			".c":          true,
			".cc":         true,
			".conf":       true,
			".config":     true,
			".cpp":        true,
			".cs":         true,
			".css":        true,
			".csv":        true,
			".Dockerfile": true,
			".gitignore":  true,
			".gitmodules": true,
			".go":         true,
			".h":          true,
			".htm":        true,
			".html":       true,
			".iml":        true,
			".js":         true,
			".json":       true,
			".jsx":        true,
			".less":       true,
			".lock":       true,
			".log":        true,
			".Makefile":   true,
			".md":         true,
			".mod":        true,
			".php":        true,
			".py":         true,
			".rb":         true,
			".rs":         true,
			".scss":       true,
			".sql":        true,
			".sum":        true,
			".svg":        true,
			".toml":       true,
			".ts":         true,
			".tsv":        true,
			".tsx":        true,
			".txt":        true,
			".xml":        true,
			".yaml":       true,
			".yml":        true,
		},
		PlainFiles: map[string]bool{
			"Dockerfile":  true,
			"license-mit": true,
			"LICENSE-MIT": true,
			"license":     true,
			"LICENSE":     true,
			"Makefile":    true,
			"readme":      true,
			"Readme":      true,
			"ReadMe":      true,
			"README":      true,
		},
		AllowedLicenseFiles: map[string]bool{
			"license-mit": true,
			"LICENSE-MIT": true,
			"license.md":  true,
			"LICENSE.md":  true,
			"license.txt": true,
			"LICENSE.txt": true,
			"LICENSE":     true,
		},
		AllowedReadMeFiles: map[string]bool{
			"readme.md":  true,
			"Readme.md":  true,
			"ReadMe.md":  true,
			"README.md":  true,
			"readme.txt": true,
			"README.txt": true,
			"README":     true,
		},
	}
}

type RepoData struct {
	Name            string
	GitURL          string
	Description     string
	BaseURL         string
	ReadMePath      string
	LicenseFilePath string
}
