package vscode

func CommandsVscode() []string {
	output := []string{
		"/usr/bin/yum update -y",
		"/usr/bin/yum install -y golang",
		"/usr/bin/yum install -y wget",
		"/usr/bin/yum install -y tar",
		"wget https://go.dev/dl/go1.20.2.linux-amd64.tar.gz",
		"tar -C /usr/local -xzf go1.20.2.linux-amd64.tar.gz",
		"rm go1.20.2.linux-amd64.tar.gz",
		"alternatives --remove go /usr/lib/golang/bin/go",
		"alternatives --install /usr/bin/go go /usr/local/go/bin/go  1000",
		"curl -fsSL https://code-server.dev/install.sh | sh",
	}
	return output
}
