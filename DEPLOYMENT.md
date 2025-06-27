# Queue Management System - DigitalOcean Deployment Guide

## Prerequisites
1. GitHub account with your code repository
2. DigitalOcean account
3. Docker installed locally (for testing)

## Deployment Steps

### Step 1: Push your code to GitHub
1. Create a new repository on GitHub (e.g., `queue-management-system`)
2. Push your code to the repository:
```bash
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/queue-management-system.git
git push -u origin main
```

### Step 2: Deploy to DigitalOcean App Platform

#### Option A: Using DigitalOcean Control Panel (Recommended)
1. Go to https://cloud.digitalocean.com/apps
2. Click "Create App"
3. Select "GitHub" as source
4. Choose your repository: `queue-management-system`
5. Select branch: `main`
6. Auto-deploy: Enable
7. DigitalOcean will automatically detect the Dockerfile
8. Review the settings:
   - App name: `queue-management-system`
   - Region: Choose your preferred region (e.g., New York - NYC)
   - Instance size: Basic ($5/month)
9. Click "Create Resources"

#### Option B: Using doctl CLI
1. Install doctl CLI tool
2. Authenticate: `doctl auth init`
3. Deploy using the app spec:
```bash
doctl apps create --spec .do/app.yaml
```

### Step 3: Environment Configuration
The app is configured to:
- Run on port 8080 (DigitalOcean standard)
- Auto-scale based on traffic
- Include Node.js dependencies for WebSocket functionality
- Serve static files from the ui/static directory

### Step 4: Access Your Application
- Your app will be available at: `https://queue-management-system-xxxxx.ondigitalocean.app`
- Default accounts are set up in your application code

## File Structure for Deployment
```
├── Dockerfile                 # Container configuration
├── .dockerignore             # Files to exclude from Docker build
├── package.json              # Node.js dependencies
├── .do/app.yaml             # DigitalOcean app specification
└── cmd/web/main.go          # Modified to support PORT env variable
```

## Cost Estimation
- Basic plan: $5/month for 512MB RAM, 1 vCPU
- Professional plans available for higher traffic

## Monitoring and Logs
- Access logs through DigitalOcean App Platform dashboard
- Monitor performance and resource usage
- Set up alerts for downtime or errors

## Custom Domain (Optional)
1. In the App Platform dashboard, go to Settings > Domains
2. Add your custom domain
3. Configure DNS records as instructed

## SSL Certificate
- Automatically provided by DigitalOcean App Platform
- Includes automatic renewal

## Troubleshooting
- Check application logs in the DigitalOcean dashboard
- Ensure all environment variables are set correctly
- Verify that the PORT environment variable is being used
- Check that static files are being served correctly
