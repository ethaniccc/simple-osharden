package script

type Firewall struct {
}

func (f *Firewall) Run() error {
	logger.Info("Installing the UFW firewall...")
	CreateCommand("apt", "update").Run()
	CreateCommand("apt", "install", "ufw").Run()

	return nil
}
