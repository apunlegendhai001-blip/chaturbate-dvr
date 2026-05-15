# Channel Persistence Solution for GitHub Actions

Your app already has built-in persistence via `SaveConfig()` and `LoadConfig()` that saves channels to `./conf/channels.json`. The issue is this file is lost between 5-hour workflow cycles.

## Solution Options

### Option 1: Use GitHub Actions Cache (Recommended - Simple)
**Pros:** Easy to implement, works with GitHub-hosted or self-hosted runners
**Cons:** Cache can be evicted if not used for 7 days

Update your workflow to cache the `conf/` directory:

```yaml
- uses: actions/cache@v4
  with:
    path: conf/
    key: dvr-channels-${{ github.run_id }}
    restore-keys: |
      dvr-channels-
```

### Option 2: Use Persistent Directory on Self-Hosted Runner
**Pros:** Most reliable, persists indefinitely
**Cons:** Requires setup on self-hosted runner

1. Create a persistent directory on your self-hosted runner:
   ```bash
   mkdir -p /mnt/dvr-data/conf
   chmod 777 /mnt/dvr-data/conf
   ```

2. Update your workflow to use a bind mount:
   ```yaml
   - name: Run DVR for 5 hours
     run: |
       mkdir -p /mnt/dvr-data/conf
       cp -r conf/ /mnt/dvr-data/conf/ 2>/dev/null || true
       nohup go run . > dvr_output.log 2>&1 &
       PID=$!
       sleep 18000
       kill $PID
       cp -r ./conf/* /mnt/dvr-data/conf/ 2>/dev/null || true
   ```

3. Or modify your Go app to use environment variable:
   ```bash
   export DVR_CONFIG_PATH=/mnt/dvr-data/conf
   ```

### Option 3: Use GitHub Actions Artifacts + Upload/Download
**Pros:** Simple, works everywhere
**Cons:** Adds execution time, size limits

Upload after each run, download at the start:
```yaml
- name: Download previous config
  uses: actions/download-artifact@v4
  with:
    name: dvr-config
    path: conf/
  continue-on-error: true

# ... run DVR ...

- name: Upload config
  uses: actions/upload-artifact@v4
  if: always()
  with:
    name: dvr-config
    path: conf/channels.json
    retention-days: 30
```

## Recommended Implementation

Use **Option 1 (Cache)** + **Option 2 (Persistent Dir)**:

1. First, add cache to your workflow (simple fallback)
2. Also add the persistent directory path on your runner (most reliable)
3. Update the Go app to support a config path via environment variable

This gives you both short-term (cache) and long-term (persistent dir) persistence.
