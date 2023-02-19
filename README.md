# firemage is a CLI tool to manage Firebase Auth Claims, written in Golang

## Overview
Let's assume you have a
- Firebase project where you're managing users in Firebase Auth.
- You need special permissions, eg. `admin`, `writer`, `editor`, `consultant`, etc.
- You're using Firestore
  1. If you're storing user data in Firestore, you can search by user properties.
  1. A list of all users with special permissions is saved as a cache. Otherwise you'd have to iterate all Firebase Auth users every time.

## Use cases
- Search users by name or email address (if you have those in your Firestore), and change their permissions.
- List all privileged users from the Firestore cache, and change their permissions.
- Save your changes to Firebase Auth and the Firestore cache.
- In case your Firestore cache and Auth Claims get out of sync, you can refresh the cache.

## Init
1. [Install Go](https://go.dev/doc/install) if you don't have it already.
1. Run `go get github.com/vendelin8/firemage`.
1. Go to https://console.cloud.google.com/iam-admin/serviceaccounts?project=YOUR_PROJECT_ID to create a service account key, and download it somewhere inside `$GOPATH/src/github.com/vendelin8/firemage` folder. 
1. Fill in `customization/custom.go` with your details
1. Optionally install [Taskfile](https://taskfile.dev/)
1. Choose language and set with `task setlang LANG`