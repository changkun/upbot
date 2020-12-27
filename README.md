# upbot

an uptime monitoring service for changkun.de

## Usage

Setup monitors in [config.yml](./configs/config.yml), then:

```
$ docker network create traefik_proxy
$ make build && make up
```

- The upbot will send notification email to the configured recipients 
  if configured monitor status is changed.
- The upbot service status can be checked from router [/upbot](https://changkun.de/upbot).

## License

MIT | Copyright &copy; 2020 [Changkun Ou](https://changkun.de)