<p align="center"> 
    <img alt="alienos" src="./static/images/logo.png" width="150" height="150" />
</p>

<h1 align="center">
Alienos
</h1>

<br/>


The Alienos is a Nostr stack (relay/blossom mediaserver/nip-05 server) which is manageable (using nip-86) and [plugin-able](wip). We designed it for self-hosting and backups.

This project is based on [Khatru](https://github.com/fiatjaf/khatru), [Event Store](https://github.com/fiatjaf/eventstore), [Blob Store](github.com/kehiy/blobstore) and [Go Nostr](github.com/nbd-wtf/go-nostr).


## Landing Page

<img alt="alienos" src="./static/images/screenshot.png" />

## Features

- [X] Support NIPs: 1, 9, 11, 40, 42, 50, 56, 59, 70, 86.
- [X] Support BUDs: 1, 2, 4, 6, 9 (Manageable using nip-86).
- [X] NIP-05 server (Manageable using nip-86, caching for recent requests to enhance response delay).
- [X] Manageable using NIP-86.
- [X] Landing page with NIP-11 document.
- [X] S3 backups (relay dbs/blobs/nip05 data/management info).
- [X] Moderator notifications.
- [X] S3 as blossom target.
- [X] Colorful Console/File logger.
- [ ] Running on Tor.
- [ ] Support plugins.
- [ ] StartOS support.
- [ ] Umbrel support.

## How to set it up?


#### **Option 1: Use Prebuilt Docker Image (Recommended)**

The easiest way to run the Alienos is by using the prebuilt image:

1. **Pull the latest image**

   ```sh
   docker pull dezhtech/alienos
   ```

2. **Run Alienos with environment variables**
   ```sh
   docker run -d --name alienos \
   -p 7771:7771 \
    -e ALIENOS_WORK_DIR="alienos_wd/" \
    -e ALIENOS_RELAY_NAME="Alienos" \
    -e ALIENOS_RELAY_ICON="https://nostr.download/6695de4b095cd99ee7b4f6e2ef9ff89a9029efc1a017e60b8b5b5cb446b2c1e0.webp" \
    -e ALIENOS_RELAY_BANNER="https://nostr.download/5b3fa3e40365061d58946fdb1bc6549a4675186591f9f589f9983895bfac8940.webp" \
    -e ALIENOS_RELAY_DESCRIPTION="A self-hosting Nostr stack!" \
    -e ALIENOS_RELAY_PUBKEY="badbdda507572b397852048ea74f2ef3ad92b1aac07c3d4e1dec174e8cdc962a" \
    -e ALIENOS_RELAY_CONTACT="hi@dezh.tech" \
    -e ALIENOS_RELAY_SELF="" \
    -e ALIENOS_RELAY_PORT=7771 \
    -e ALIENOS_RELAY_BIND="0.0.0.0" \
    -e ALIENOS_RELAY_URL="" \
    -e ALIENOS_BACKUP_ENABLE="true" \
    -e ALIENOS_BACKUP_INTERVAL_HOURS=1 \
    -e ALIENOS_S3_ACCESS_KEY_ID="" \
    -e ALIENOS_S3_SECRET_KEY="" \
    -e ALIENOS_S3_ENDPOINT="" \
    -e ALIENOS_S3_REGION="" \
    -e ALIENOS_S3_BUCKET_NAME="alienos" \
    -e ALIENOS_S3_AS_BLOSSOM_STORAGE="false" \ 
    -e ALIENOS_S3_BLOSSOM_BUCKET="alienos" \
    -e ALIENOS_PUBKEY_WHITE_LISTED="false" \
    -e ALIENOS_KIND_WHITE_LISTED="false" \
    -e ALIENOS_ADMINS="" \
    -e ALIENOS_LOG_FILENAME="alienos.log" \
    -e ALIENOS_LOG_LEVEL="info" \
    -e ALIENOS_LOG_TARGETS="file,console" \
    -e ALIENOS_LOG_MAX_SIZE=10 \
    -e ALIENOS_LOG_FILE_COMPRESS=true \
   dezhtech/alienos
   ```

---

#### **Option 2: Using Docker Compose**

For a more structured deployment, use **Docker Compose**:

1. **use `compose.yml`**
use the exist compose file in the alienos directory

2. **Run with Compose**
   ```sh
   docker-compose up -d
   ```

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
