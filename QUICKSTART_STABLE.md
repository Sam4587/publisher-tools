# Quick Start Guide - Stable Service Manager

## Problem Solved

The original `start-stable.bat` had encoding issues with Chinese characters in Windows CMD. Now it's rewritten in English, following the perfect format of `start.bat`.

## How to Use

### Option 1: Double-click `start-stable.bat`
- Simple and easy
- Starts service manager automatically
- Opens browser automatically

### Option 2: Manual start
```bash
node service-manager.js
```

## What's Different from start.bat?

| Feature | start.bat | start-stable.bat |
|---------|-----------|------------------|
| Service Monitoring | ❌ No | ✅ Yes (every 30s) |
| Auto Restart | ❌ No | ✅ Yes (5s delay) |
| Health Check | ❌ No | ✅ Yes |
| Error Handling | ⚠️ Basic | ✅ Advanced |
| Log Management | ⚠️ Basic | ✅ Complete |

## Service Manager Features

1. **Auto Monitoring**
   - Checks service status every 30 seconds
   - Displays PID, port, uptime, and health status

2. **Auto Restart**
   - Automatically restarts crashed services
   - 5-second delay before restart
   - Prevents service downtime

3. **Health Check**
   - HTTP health check on `/api/health` endpoint
   - Detects service failures early
   - Ensures service availability

4. **Complete Logging**
   - All logs saved to `logs/` directory
   - Timestamped entries
   - Error tracking

## Service Architecture

```
Service Manager (Node.js)
    ├── Node.js Backend (Port 3001)
    │   └── API Server
    └── Frontend (Port 5173)
        └── Vite Dev Server
```

## Troubleshooting

### Issue: Service won't start
**Solution:**
1. Check if ports are in use:
   ```bash
   netstat -ano | findstr ":3001"
   netstat -ano | findstr ":5173"
   ```

2. Kill processes if needed:
   ```bash
   stop.bat
   ```

3. Restart services:
   ```bash
   start-stable.bat
   ```

### Issue: Service keeps restarting
**Solution:**
1. Check logs in `logs/` directory
2. Verify dependencies:
   ```bash
   cd server && npm install
   cd publisher-web && npm install
   ```

### Issue: API returns 500 error
**Solution:**
1. Check if backend is running:
   ```bash
   curl http://localhost:3001/api/health
   ```

2. Check service manager status
3. Review logs for errors

## Comparison: Before vs After

### Before (start.bat)
- Services run independently
- No monitoring
- Manual restart required on crash
- Limited error handling

### After (start-stable.bat)
- Centralized service management
- Continuous monitoring
- Automatic restart on failure
- Comprehensive error handling
- Health check mechanism

## Recommended Usage

**Development:**
- Use `start-stable.bat` for better stability
- Monitor service status in console
- Check logs when issues occur

**Production:**
- Use PM2 for process management
- Or use Docker containers
- Set up monitoring alerts

## Files Created

1. `start-stable.bat` - Stable startup script (English)
2. `service-manager.js` - Service manager with monitoring
3. `SERVICE_STABILITY.md` - Detailed documentation
4. `QUICKSTART_STABLE.md` - This quick start guide

## Next Steps

1. Test the stable startup:
   ```bash
   start-stable.bat
   ```

2. Verify services are running:
   - Frontend: http://localhost:5173
   - Backend: http://localhost:3001/api/health

3. Monitor service status in the Service Manager window

4. Check logs if needed:
   - `logs/node-backend.log`
   - `logs/frontend.log`

## Summary

The new `start-stable.bat` provides:
- ✅ No encoding issues (English only)
- ✅ Automatic service monitoring
- ✅ Auto-restart on failure
- ✅ Health checks
- ✅ Complete logging
- ✅ Better stability

Your services will now run much more reliably!
