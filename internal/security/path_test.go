package security
import "testing"
func TestValidID(t *testing.T){for _,id:=range []string{"../../etc/passwd","a/b","a\\b"}{if ValidID(id){t.Fatalf("accepted %q",id)}};if !ValidID("betrieb:sap:backup"){t.Fatal("valid id rejected")}}
