package facts

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type Fact struct {
	Value      string     `json:"value"`
	Evidence   string     `json:"evidence"`
	Confidence Confidence `json:"confidence"`
}

type Facts struct {
	Runtime        Fact              `json:"runtime"`
	Framework      Fact              `json:"framework"`
	PackageManager Fact              `json:"package_manager"`
	NodeVersion    Fact              `json:"node_version"`
	Scripts        map[string]string `json:"scripts"`
	EnvFiles       []string          `json:"env_files"`
	Warnings       []string          `json:"warnings"`
}

func New() *Facts {
	return &Facts{
		Scripts:  make(map[string]string),
		EnvFiles: make([]string, 0),
		Warnings: make([]string, 0),
	}
}

func (f *Facts) AddWarning(warning string) {
	if warning == "" {
		return
	}
	f.Warnings = append(f.Warnings, warning)
}
