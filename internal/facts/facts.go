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
	Type                  Fact              `json:"type"`
	Runtime               Fact              `json:"runtime"`
	Framework             Fact              `json:"framework"`
	SecondaryFrameworks   []Fact            `json:"secondary_frameworks,omitempty"`
	PackageManager        Fact              `json:"package_manager"`
	PackageManagerVersion string            `json:"package_manager_version,omitempty"`
	NodeVersion           Fact              `json:"node_version"`
	Lockfile              string            `json:"lockfile,omitempty"`
	Scripts               map[string]string `json:"scripts"`
	EnvFiles              []string          `json:"env_files"`
	Warnings              []string          `json:"warnings"`
}

func New() *Facts {
	return &Facts{
		SecondaryFrameworks: make([]Fact, 0),
		Scripts:             make(map[string]string),
		EnvFiles:            make([]string, 0),
		Warnings:            make([]string, 0),
	}
}

func (f *Facts) AddWarning(warning string) {
	if warning == "" {
		return
	}
	f.Warnings = append(f.Warnings, warning)
}
