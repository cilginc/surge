- Pausing downloads(done)
- Duplicate downloads
- Implement filename in TUI
- add filters/sort by in TUI?
- better colors in TUI?
- Add ability to browse locations during adding download
- List completed downloads feature
- Add backoff of connections incase server sends too many requests.
- Tune Chunk Size Tune number of connections(How to add when to add)?
- Add support for multiple links or mirrors
- Add support for bit torrent?
- ./surge directory is currently local should be global in prod
- We should design each page with much better colors and intuition
- Check everything works with the extension properly
- TUI gets completely stuck on a large number of downloads
Use uuid
add .surge to incomplete files
Fix pause restarts at 60% after pausing at 90%
Pausing hash check and download hash check
make surge server default(choose a port if already used)

clipboard monitoring


Fix dead link thing
No crash recovery for active downloads. Only explicitly paused downloads are persisted. 
Power failure = restart from zero.O(n) file I/O on every operation via JSON master list. Will cause visible lag with large download history.

Add authentication to HTTP server — At minimum, generate a random token on startup and require it in headers.

Implement checksum verification — Wire up the existing flags. Verify after download completes, delete file if mismatch.

Write a real README — Installation, usage, features, screenshots, comparison table, contributing guide.

Split large files — 
concurrent.go
 → 
concurrent.go
, task_queue.go, work_stealing.go, health_monitor.go. Same for 
update.go
.

Add integration tests — Spin up a local HTTP server, download test files, verify checksums, test pause/resume.