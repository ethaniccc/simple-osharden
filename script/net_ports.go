package script

// PortListeningApplications is a script that checks for any applications that are listening on ports.
// The user will confirm that these applications are allowed to listen on these ports, and if they're
// not, the process will be killed. The user will also be prompted if the program should be
// removed from the machine.
type PortListeningApplications struct {
}

func (r *PortListeningApplications) Name() string {
	return "apps-on-ports"
}

func (r *PortListeningApplications) Description() string {
	return "Checks for any applications that are listening on ports."
}

func (r *PortListeningApplications) Run() {
	// TODO: Implementation.
}
