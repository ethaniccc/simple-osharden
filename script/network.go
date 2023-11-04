package script

type Firewall struct {
}

func (f *Firewall) Name() string {
	return "Firewall"
}

func (f *Firewall) Description() string {
	return "Installs UFW and configures it."
}

func (f *Firewall) Run() error {
	return ExecuteLoggedCommands([]LoggedCommand{
		{"Installing UFW", "apt install ufw", true},
		{"Enabling UFW Firewall", "ufw enable", true},
		{"Allowing SSH through firewall", "ufw allow ssh", false},
		{"Setting option to reject incoming connections by default", "ufw default reject incoming", false},
		{"Setting option to allow outgoing connections by default", "ufw default allow outgoing", false},
	})
}
