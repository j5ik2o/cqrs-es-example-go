# Tools Installation

## Go

```shell
$ curl -sLk https://raw.githubusercontent.com/kevincobain2000/gobrew/master/git.io.sh | sh
$ export PATH="$HOME/.gobrew/current/bin:$HOME/.gobrew/bin:$PATH"
```

https://github.com/kevincobain2000/gobrew

## asdf

```sh
$ brew install asdf
$ echo -e "\n. $(brew --prefix asdf)/libexec/asdf.sh" >> ${ZDOTDIR:-~}/.zshrc
```

https://asdf-vm.com/

For Ubuntu, install according to the following

https://asdf-vm.com/guide/getting-started.html

## jq

```shell
$ asdf plugin-add jq https://github.com/AZMCode/asdf-jq.git
$ asdf install jq 1.6
$ asdf local jq 1.6
$ jq --version
jq-1.6
```

https://github.com/ryodocx/asdf-jq

### Docker

#### Mac

Follow the instructions below to install it.

Docker Desktop for Mac

https://docs.docker.com/desktop/install/mac-install/

#### Ubuntu

Follow the instructions below to install it.

- Docker Desktop for Ubuntu
  - https://docs.docker.com/desktop/install/debian/
- Docker Engine for Ubuntu
  - https://docs.docker.com/engine/install/ubuntu/

In addition, to be able to use it without sudo, please follow these steps.

https://qiita.com/katoyu_try1/items/1bdaaad9f64af86bbfb7
