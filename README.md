## Firemage
![Coverage](https://img.shields.io/badge/Coverage-60.3%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/vendelin8/firemage)](https://goreportcard.com/report/github.com/vendelin8/firemage)

firemage is a CLI tool to manage Firebase Auth Claims, written in Golang

The project uses `task` to make your life easier. If you're not familiar with Taskfiles you can take a look at [this quickstart guide](https://taskfile.dev/).

## Overview
Let's assume you have a
- Firebase project where you're managing users in Firebase Auth.
- You need special permissions, eg. `admin`, `writer`, `editor`, `consultant`, etc.
- You're using Firestore
  1. If you're storing user data in Firestore, you can search by user properties.
  1. A list of all users with special permissions is saved as a cache. Otherwise you'd have to iterate all Firebase Auth users every time.

## Use cases
- List all privileged users from the Firestore cache.
- Search users by name or email address (if you have those in your Firestore).
- Edit permissions of listed or searched users.
- Save permission changes to Firebase Auth and the Firestore cache.
- In case your Firestore cache and Auth Claims get out of sync, you can refresh the cache.

## Setup
1. [Install Go](https://go.dev/doc/install) if you don't have it already.
1. Run:

```bash
go get github.com/vendelin8/firemage
```

1. Go to https://console.cloud.google.com/iam-admin/serviceaccounts?project=YOUR_PROJECT_ID to create a service account key, and download it somewhere inside `$GOPATH/src/github.com/vendelin8/firemage` folder. The default path is `service-account.json`, change it in `Taskfile` if you want it otherwise.
1. Fill in `custom/custom.txt` with your details. These will be built into the binary.
1. Localization will be built into the binary too. The default is English (`LANG`=`en`). If you want to change it to your language, and you can find it in `i18n` folder, call:

```bash
task setlang LANG=<LANG>
```

## Configurate keyboard shortcuts
You can overwrite the defaults by editing `conf.yml`. It's localized with `task setlang`, see above. You can define more shortcuts to functions as well.

## Help
You can print help with

```bash
task help
```

## Debug
If you have Taskfile, you can call it with:

```bash
task debug
```

To add arguments here, do like `task debug -- -v`.
Otherwise with `go run .` or similar with your own args. Or from any path with something like `go run github.com/vendelin8/firemage`.
To print debug info to `log.txt` add `-v`.
To use it with a configured FIrebase emulator add `-e`.

## Build
You can compile with

```bash
task build
```

or cross compile to multiple platforms with `task build-win`, `task build-osx` or `task build-lin`. It will output to `build` folder. It compiles `custom/custom.go` options and chosen language. You can ship with the compiled version to a teammate. Add `service-account.json` in the same folder and optionally `conf.yml` to be able to configure keyboard shortcuts.

## How Firestore caching works
The first `Refresh` call will create a collection `misc` with a document `specialUsers`. It will have all privileged users as `uid` -> `email` pairs as data. When you open the `List` page in the app, it will download this list, and get the permissions from Firebase Auth claims. By removing permissions and calling `Save` users may be removed from the cache list. By searching for email or name, adding permissions to other users and calling `Save`, users may be added to the cache list.

## Test & lint

Run linting

```bash
task lint
```

Run tests

```bash
task test
```

## Contribute
Feel free to raise an issue or create a PR, eg. translate the package to your own language.
