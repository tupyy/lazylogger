package conf

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestConf(t *testing.T) {
	conf := `[hosts]
    [hosts.A]
        address = "192.168.2.1"
        username = "cosmin"
        key = "supersupersecretkey"
    [hosts.B]
        address = "192.168.2.1"
        username = "cosmin"
        key = "supersupersecretkey"

[datasources]
    [datasources.d1]
    host = "a"
    jumpHost = "b"
    file = "test.file"

[aliases]
    [aliases.alias1]
    datasource = "d1"
    command = "test"
    flags = "-f -g"
`
	content := []byte(conf)

	tmpfn := "temp_conf_file.toml"
	if err := ioutil.WriteFile(tmpfn, content, 0666); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfn)

	cc, err := ReadConfiguration(tmpfn)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v", cc)
}
