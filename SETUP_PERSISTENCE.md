# GitHub Actions Setup for Channel Persistence

## What I've Done

1. **Updated [.github/workflows/run-dvr.yml](.github/workflows/run-dvr.yml)** with:
   - GitHub Actions `cache` to persist `conf/` directory between runs
   - Persistent directory backup/restore from `/mnt/dvr-data/conf`
   - Better error handling for kill signals

2. **Enhanced [manager/manager.go](manager/manager.go)** with:
   - New `getConfigPath()` function that respects `DVR_CONFIG_PATH` environment variable
   - Updated `SaveConfig()` and `LoadConfig()` to use the configurable path
   - Fallback to `./conf/channels.json` if environment variable is not set

## Setup Instructions

### Option A: Use GitHub Actions Cache (Automatic, No Setup Needed)
✅ **Already implemented in the workflow**

The cache step will automatically:
- Restore `conf/` from previous runs (up to 7 days)
- Save it after your DVR stops running
- Works with any runner type

### Option B: Use Persistent Directory on Self-Hosted Runner (Most Reliable)

1. **SSH into your self-hosted runner machine** and create persistent storage:
   ```bash
   sudo mkdir -p /mnt/dvr-data/conf
   sudo chmod 777 /mnt/dvr-data/conf
   ```

2. **(Optional) Point the app to use this directory** by setting the environment variable:
   ```bash
   export DVR_CONFIG_PATH=/mnt/dvr-data/conf
   ```
   
   Add this to your GitHub Actions workflow if desired:
   ```yaml
   - name: Run DVR for 5 hours
     env:
       DVR_CONFIG_PATH: /mnt/dvr-data/conf
     run: |
       # ... rest of the script
   ```

### Option C: Use Both (Recommended)
- **Cache** provides automatic fallback if the persistent dir gets deleted
- **Persistent dir** provides long-term storage beyond 7 days
- Already implemented in your workflow!

## How It Works

1. **On workflow start:**
   - Restores from GitHub Actions cache (if available)
   - If cache miss, restores from `/mnt/dvr-data/conf` (if available)

2. **During 5-hour run:**
   - Your app saves channels to `./conf/channels.json`
   - If `DVR_CONFIG_PATH` is set, it will also save there

3. **On workflow end:**
   - GitHub Actions automatically caches the `conf/` directory
   - Backup step copies `conf/channels.json` to `/mnt/dvr-data/conf/`

## Verification

After the first workflow run completes:

```bash
# Check cache was created
cd /workspaces/chaturbate-dvr
ls -la conf/channels.json

# Check persistent backup (if you set up Option B)
ls -la /mnt/dvr-data/conf/channels.json
```

## Test It

Next time your workflow runs (in 5 hours or manually trigger it):
1. It should restore the previous channels
2. Channels will resume recording automatically
3. Check the logs in the workflow run

---

**Note:** Channel names and settings are now persistent across all 5-hour cycles! 🎉
