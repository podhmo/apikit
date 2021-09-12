package tinypkg

import "fmt"

func ToImportPackageString(ip *ImportedPackage) string {
	if ip.qualifier != "" {
		return fmt.Sprintf("%s %q", ip.qualifier, ip.pkg.Path)
	}
	return fmt.Sprintf("%q", ip.pkg.Path)
}
