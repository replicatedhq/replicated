package pkgmgr

type ExternalPackageManager interface {
	IsInstalled() (bool, error)
	UpgradeCommand() string
}
