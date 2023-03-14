package internal

//go:generate easyjson

type (
	Config struct {
		ApiKey  string   `yaml:"api_key"`
		SshCert string   `yaml:"ssh_cert"`
		Folder  string   `yaml:"folder"`
		Owners  []string `yaml:"owners"`
	}

	DataModel map[string]string

	//easyjson:json
	GitHubReposModel struct {
		Name string `json:"full_name"`
		Url  string `json:"ssh_url"`
	}

	//easyjson:json
	GitHubResponseModel struct {
		Items []GitHubReposModel `json:"items"`
	}
)
