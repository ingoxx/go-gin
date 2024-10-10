package assets

type program struct {
	Pid    int    `json:"pid"`
	Action int    `json:"action"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Kind   int    `json:"type"`
	Load   bool   `json:"load"`
}

type ProgramConfig struct {
	Config []program         `json:"config"`
	Select map[string]string `json:"select"`
}

func NewProgramConfig() *ProgramConfig {
	return &ProgramConfig{
		Config: []program{
			{
				Name:   "docker更新",
				Pid:    2,
				Action: 1,
				Value:  "dockerUpdate",
				Kind:   1,
				Load:   false,
			},
			{
				Name:   "java更新",
				Pid:    4,
				Action: 2,
				Value:  "javaUpdate",
				Kind:   1,
				Load:   false,
			},
			{
				Name:   "重启docker",
				Pid:    5,
				Action: 3,
				Value:  "dockerReload",
				Kind:   2,
				Load:   false,
			},
			{
				Name:   "重启java",
				Pid:    6,
				Action: 4,
				Value:  "javaReload",
				Kind:   2,
				Load:   false,
			},
		},
		Select: map[string]string{
			"docker更新":    "dockerUpdate",
			"java更新":      "javaUpdate",
			"重启docker":    "dockerReload",
			"重启java":      "javaReload",
			"docker更新Log": "dockerUpdateLog",
			"java更新Log":   "javaUpdateLog",
		},
	}
}
