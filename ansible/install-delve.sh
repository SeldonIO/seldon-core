#!/bin/bash

CLUSTER_NAME="seldon"
GO_VERSION="1.24.7"

echo "Installing Go and Delve in kind cluster: $CLUSTER_NAME"

NODES=$(kind get nodes --name "$CLUSTER_NAME")

for node in $NODES; do
    echo "Setting up $node..."

    docker exec "$node" bash -c "
        # Install Go
        curl -sL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz -o /tmp/go.tar.gz && \
        rm -rf /usr/local/go && \
        tar -C /usr/local -xzf /tmp/go.tar.gz && \
        rm /tmp/go.tar.gz && \

        # Install Delve
        /usr/local/go/bin/go install github.com/go-delve/delve/cmd/dlv@latest && \

        # Create symlinks
        ln -sf /usr/local/go/bin/go /usr/local/bin/go && \
        ln -sf /root/go/bin/dlv /usr/local/bin/dlv && \

        # Verify
        echo '=== Versions ===' && go version && dlv version
    "

    echo "✓ Setup complete on $node"
done