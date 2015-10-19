## Supported clients

You should be able to use any client that supports memcached TCP text protocol.
Reliable reads would only work if client maintains persistent connection to the server
between get requests.

Here is a list of clients that would support reliable reads:

## Ruby
- [kestrel-client](https://github.com/freels/kestrel-client)
- [dalli](https://github.com/mperham/dalli)

## Go
- [kklis/gomemcache](https://github.com/kklis/gomemcache)

If you find other compatible clients,
please don't hesitate to create a PR and update this list.
