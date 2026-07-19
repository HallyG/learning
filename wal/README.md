# WAL

Extremely simple Write Ahead Log (WAL).

Improvements:

- Batch flushing and fsync - amortize the expensive I/O operations.
- Crash on fsync? re Postgres fsyncgate.
- Truncate WAL if we find a checksum error.
- File locking? 2 processes could try and write to the WAL.
- Checkpointing to prevent WAL from growing forever.
