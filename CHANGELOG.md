## 0.3

- Implemented GET <queue-name>/close/open
- Accept /t=<milliseconds> parameter for backwards compatibility.
  Ignore provided timeout.

## 0.2.2

- Fixed bug that was introduced in commit 95912a4.
  Did not handle EOF disconnects properly.

## 0.2.1

- Allow uppercase commands
- Add open_transactions to stats
- Use testify/assert for tests
