<p align="center"> 
    <img alt="alienos" src="./static/images/logo.png" width="150" height="150" />
</p>

<h1 align="center">
Alienos
</h1>

<br/>


The Alienos is a Nostr stack (relay/blossom mediaserver/nip-05 server) which is manageable (using nip-86) and [plugin-able](wip). We designed it for self-hosting and backups.

This project is based on [Khatru](https://github.com/fiatjaf/khatru), [EventStore](https://github.com/fiatjaf/eventstore), [BlobStore](github.com/kehiy/blobstore) and [go-nostr](github.com/nbd-wtf/go-nostr).

## Features

- [X] Support NIPs: 1, 9, 11, 40, 42, 50, 56, 59, 70, 86.
- [X] Support BUDs: 1, 2, 4, 6, 9.
- [X] NIP-05 server.
- [X] Manageable using NIP-86.
- [X] Landing page with NIP-11 document.
- [X] S3 backups (relay dbs/blobs/nip05 data/management info).
- [ ] Full setup document.
- [ ] Moderator notifications.
- [ ] StartOS support.
- [ ] Umbrel support.
- [ ] Support plugins.

## How to set it up?

### VPS

> TODO.

#### Docker

> TODO.

#### OS

> TODO.

### Umbrel

> TODO.

### StartOS

> TODO.

## Limitations

This project is highly suitable for personal, community, team and backup usage since its light-weight, feature-full and easy to setup/manage.

If you are aiming to run a relay/nip-05 server/blossom media server for large scale and high load (as a paid relay, default relay fo your client or a public global relay) you can consider using the [Immortal](https://github.com/dezh-tech/immortal) relay and its adjacent projects.

## Contribution

All kinds of contributions are welcome!

## Donation

Donations and financial support for the development process are possible using Bitcoin and Lightning:

**on-chain**:

```
bc1qa0z44j7m0v0rx85q0cag5juhxdshnmnrxnlr32
```

**lightning**: 

```
dezh@coinos.io
```
