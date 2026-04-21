package cli

var (
	version   = "dev"
	buildTime = "unknown"
)

type app struct {
	configFile string
	output     string
}

type config struct {
	Tower towerConfig `mapstructure:"tower"`
	User  userConfig  `mapstructure:"user"`
}

type towerConfig struct {
	Host               string `mapstructure:"host"`
	Scheme             string `mapstructure:"scheme"`
	InsecureSkipVerify bool   `mapstructure:"insecureSkipVerify"`
}

type userConfig struct {
	Name     string `mapstructure:"name"`
	Password string `mapstructure:"password"`
	Source   string `mapstructure:"source"`
}

type vmQueryOptions struct {
	Name string
	All  bool
}

type createVMOptions struct {
	Name          string
	Cluster       string
	Template      string
	IsFullCopy    bool
	VCPU          int
	MemoryGB      int
	HA            bool
	Status        string
	Firmware      string
	Bus           string
	ConnectVlanID string
	VMVolumeID    string
	ElfImageID    string
	DiskSizeGB    int
	DiskName      string
	Password      string
	Hostname      string
	UserData      string
	Nameservers   []string
	PublicKeys    []string
	IPAddress     string
	Netmask       string
	NicIndex      int
}
