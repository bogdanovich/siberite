## 0.6.4 (unreleased)
- Added PID file support (thanks to @jeteon)

## 0.6.3
- Added support for 'quit' command (memcached protocol compatibility)

## 0.6.2
- Limit OpenFilesCacheCapacity to 64 per queue

## 0.6.0
- Change durable cursor separator from `:` to `.` for Kestrel compatibility.<br>
  Kestrel uses `:` as a namespace separator.
- Allow `:` character in queue name

## 0.5.2
- Accept connections only after queues are fully initialized

## 0.5.1

- Fanout queues support.<br>
  `set <queue>+<another_queue>+<third_queue> ...` adds an item to multiple queues.

## 0.5

- Add durable cursors. An ability to consume queue multiple times
  using `get <queue>:<cursor>` syntax

- New directory structure is backwards incompatible with  v0.4.x.
  But you can manually move each existing `data/<queue>` directory to
  `data/<queue>/<queue>` and new siberite will pick up the data.

## 0.4.2

- Enable leveldb BlockCacher (improves performance)

## 0.4.1

- Fix repository.GetQueue returns error without lock release

## 0.4

- Add GETS command support (for protocol compatibility)

## 0.3.1

- Fix deadlock during FLUSH_ALL
- Fix race condition when opening a new queue

## 0.3

- Implement GET <queue-name>/close/open
- Accept /t=<milliseconds> parameter for backwards compatibility.
  Ignore provided timeout.

## 0.2.2

- Fix bug that was introduced in commit 95912a4.
  Did not handle EOF disconnects properly.

## 0.2.1

- Allow uppercase commands
- Add open_transactions stats parameter
- Use testify/assert for tests
