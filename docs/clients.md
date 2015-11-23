## Supported clients

You should be able to use any client that supports memcached TCP text protocol.
Reliable reads would only work if client maintains persistent connection to the server
between get requests.

Here is a list of clients that would support reliable reads:

## Ruby
- [siberite-client](https://github.com/bogdanovich/siberite-ruby)
- [memcache-client](https://github.com/mperham/memcache-client)

## Go
- [kklis/gomemcache](https://github.com/kklis/gomemcache)

If you find other compatible clients,
please don't hesitate to create a PR and update this list.
