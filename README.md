# home-dns

Use [Route53](https://aws.amazon.com/route53/) for dynamic DNS.

Reads record name and zone from a YAML config and upserts on schedule.

## Usage

### Build

```shell
go get
go build
```

### Config

The config file should have a `source` endpoint and each record to be updated:

```yaml
resource: "https://api.ipify.org?format=text"

records:
- cron: "@hourly"
  name: home.example.com.
  zone: Z1ABCDEF123456
- cron: "@daily"
  name: office.example.com.
  zone: Z1ABCDEF123456
```

### Run

Execute with the config file:

```shell
AWS_PROFILE="home-root" ./home-dns config.yml
```

## Features

- multiple zones and records
- update on cron schedule

### Roadmap

- cache the external address (on a schedule of its own?)
- interface selection for external check
