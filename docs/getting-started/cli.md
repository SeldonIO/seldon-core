---
description: >-
  Learn about installing Seldon command line tool that you can use to manage
  Seldon Core 2 resources.
---

# Seldon CLI

{% hint style="info" %}
**Note**: You can use Seldon CLI and `kubectl` for managing Seldon Core 2 resources. You can choose the appropriate command line tool based on how Seldon CLI and `kubectl` handle control plane and data plane operations. For more details, see [Seldon CLI](../cli/).
{% endhint %}

## Installing Seldon CLI

To install Seldon CLI using prebuild binaries or build them locally.

{% tabs %}
{% tab title="Linux binary" %}
1. Download from a recent release from `https://github.com/SeldonIO/seldon-core/releases`.\
   It is dynamically linked and will require and \*nix architecture and glibc 2.25+.
2.  Move to the `seldon` folder and provide the permissions.

    ```
    mv seldon-linux-amd64 seldon
    chmod u+x seldon
    ```
3. Add the folder to your PATH.
{% endtab %}

{% tab title="Local build" %}
1. Install Go version `1.21.1`
2.  Clone and make the build.

    ```bash
    git clone https://github.com/SeldonIO/seldon-core --branch=v2
    cd seldon-core/operator
    make build-seldon
    ```
3. Add `<project-root>/operator/bin` to your PATH.
{% endtab %}

{% tab title="macOS ARM build" %}
1. Install dependencies.\
   `brew install go librdkafka`
2.  Clone the repository and make the build.

    ```
    git clone https://github.com/SeldonIO/seldon-core --branch=v2
    cd seldon-core/operator
    make build-seldon-arm
    ```
3. Add `<project-root>/operator/bin` to your PATH.&#x20;
4. Open your terminal and open up your `.bashrc` or `.zshrc` file and add the following line:\
   `export PATH=$PATH:<project-root>/operator/bin`
{% endtab %}
{% endtabs %}

## Usage

```sh
seldon --help
```
