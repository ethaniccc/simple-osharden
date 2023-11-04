package script

// Firewall is a script that installs UFW and configures it. By default, it allows SSH
// connections, and rejects all other incoming connections. All outgoing connections
// are allowed by default.
type Firewall struct {
}

func (f *Firewall) Name() string {
	return "firewall"
}

func (f *Firewall) Description() string {
	return "Installs UFW and configures it."
}

func (f *Firewall) Run() error {
	return ExecuteLoggedCommands([]LoggedCommand{
		{"Installing UFW", "apt install ufw", true},
		{"Enabling UFW Firewall", "ufw enable", false},
		{"Allowing SSH through firewall", "ufw allow ssh", false},
		{"Setting option to reject incoming connections by default", "ufw default reject incoming", false},
		{"Setting option to allow outgoing connections by default", "ufw default allow outgoing", false},
	})
}
