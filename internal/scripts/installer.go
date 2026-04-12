package scripts

type Installer interface {
	InstallPackage(pkg string) (bool, string)
	InstallParu() (bool, string)
	InstallVSCodeExtension(extension string) (bool, string)
	EnableAutologin() (bool, string)
	EnablePasswordlessSSH() (bool, string)
	EnablePasswordlessSudo() (bool, string)
	GetPackageDescription(item string) string
	CheckParuInstalled() (bool, string)
}

type Runner struct{}

func (r Runner) InstallPackage(pkg string) (bool, string) {
	return InstallPackage(pkg)
}

func (r Runner) InstallParu() (bool, string) {
	return InstallParu()
}

func (r Runner) InstallVSCodeExtension(extension string) (bool, string) {
	return InstallVSCodeExtension(extension)
}

func (r Runner) EnableAutologin() (bool, string) {
	return EnableAutologin()
}

func (r Runner) EnablePasswordlessSSH() (bool, string) {
	return EnablePasswordlessSSH()
}

func (r Runner) EnablePasswordlessSudo() (bool, string) {
	return EnablePasswordlessSudo()
}

func (r Runner) GetPackageDescription(item string) string {
	return GetPackageDescription(item)
}

func (r Runner) CheckParuInstalled() (bool, string) {
	return CheckParuInstalled()
}
