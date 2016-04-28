package luser

import (
	"bufio"
	"os"
	"strings"
)

type Version struct {
	Lustre string `yaml:"lustre"`
	Kernel string `yaml:"kernel"`
	Build  string `yaml:"build"`
}

func GetVersion() (*Version, error) {
	var ver Version

	fp, err := os.Open("/proc/fs/lustre/version")
	if err != nil {
		return nil, err
	}
	b := bufio.NewReader(fp)
	for {
		label, err := b.ReadString(':')
		if err != nil {
			return &ver, nil
		}
		value, err := b.ReadString('\n')
		if err != nil {
			return &ver, nil
		}
		value = strings.TrimSpace(value)
		switch label {
		case "build:":
			ver.Build = value
		case "kernel:":
			ver.Kernel = value
		case "lustre:":
			ver.Lustre = value
		}
	}
}