# Common Requirements

Install ``kubectl``


The Kubernetes command-line tool `kubectl` allows you to run commands against Kubernetes clusters. Follow the steps below to install it based on your OS.

### macOS

You can install `kubectl` using **Homebrew**:

```bash
brew install kubectl
```

Alternatively, using the official Kubernetes release:
```bash
curl -LO "https://dl.k8s.io/release/$(curl -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```


### Linux

Using the official Kubernetes release:
```bash
curl -LO "https://dl.k8s.io/release/$(curl -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

After installation, verify:
```bash
kubectl version --client
```
