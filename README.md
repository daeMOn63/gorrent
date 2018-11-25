# Gorrent

Gorrent's objective is to recreate a peer to peer file sharing application in go, mainly for educationnal purpose.
It's taking inspiration over the implementation of bittorrent, but getting ride of annoying bencode format and some other features on the way. The main focus is around the fast transfer of files across machines in a private environment.

## Notes

### General

#### Create gorrent
`go run gorrent.go create -src /path/to/sources -dst /tmp/some.gorrent -announce 127.0.0.1:4444`

### Trackerd

#### Launch trackerd
`go run gorrent.go trackerd`

### Peerd

#### Launch peerd
`go run gorrent.go peerd -config peerd_config_sample2.json`

#### List gorrents
`curl -XGET --unix-socket /tmp/gorrent/peerd.sock http://localhost/`

#### Add new gorrent
`curl -XPOST --unix-socket /tmp/gorrent/peerd.sock -F "gorrent=@/tmp/some.gorrent" -F "path=/path/to/storage/"  http://localhost/add`

