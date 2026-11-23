package shared

import "testing"

func TestQualifiedNamesRequireLiteralSeparators(t *testing.T) {
	for _,name:=range []string{"finance.invoice","finance.invoice.view","profiles.accountant"}{
		if !ValidQualifiedName(name){t.Fatalf("valid qualified name rejected: %q",name)}
	}
	for _,name:=range []string{"finance/invoice","finance-invoice","finance invoice","finance@invoice","finance..invoice","finance."}{
		if ValidQualifiedName(name){t.Fatalf("invalid separator accepted: %q",name)}
	}
	for _,v:=range []string{"1.0.0","12.34.567"}{
		if !ValidReleaseVersion(v){t.Fatalf("valid semver rejected: %s",v)}
	}
	for _,v:=range []string{"1a0a0","1/0/0","1-0-0","1.0.a"}{
		if ValidReleaseVersion(v){t.Fatalf("invalid version accepted: %s",v)}
	}
}
