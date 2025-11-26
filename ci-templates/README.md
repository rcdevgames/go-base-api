# CI/CD Templates for Modular Monolith Go API

This directory contains CI/CD pipeline templates for different platforms. Each template is configured for a Go application with Docker support.

## Available Templates

### 1. GitLab CI/CD
**File**: `gitlab/.gitlab-ci.yml`

**Features**:
- ✅ Multi-stage pipeline (test → build → security → deploy)
- ✅ Go testing with coverage reports
- ✅ GolangCI-Lint for code quality
- ✅ Docker build and push to GitLab Container Registry
- ✅ Security scanning with Gosec
- ✅ Manual deployment to staging/production

**Setup**:
1. Copy `.gitlab-ci.yml` to root of your repository
2. Configure GitLab Runner with Docker support
3. Set up GitLab Container Registry if needed
4. Adjust deployment commands for your infrastructure

### 2. GitHub Actions
**File**: `github-actions/ci.yml`

**Features**:
- ✅ Matrix builds with latest Ubuntu
- ✅ Go testing with Codecov integration
- ✅ GolangCI-Lint for code quality
- ✅ Security scanning with Gosec
- ✅ Docker build and push to Docker Hub
- ✅ Environment-based deployments
- ✅ Caching for faster builds

**Setup**:
1. Create `.github/workflows/` directory in your repository
2. Copy `ci.yml` to `.github/workflows/ci.yml`
3. Configure these secrets in GitHub:
   - `DOCKER_USERNAME`
   - `DOCKER_PASSWORD`
   - `CODECOV_TOKEN` (optional, for coverage reporting)
4. Adjust deployment steps for your infrastructure

### 3. Jenkins Pipeline
**File**: `jenkins/Jenkinsfile`

**Features**:
- ✅ Declarative pipeline syntax
- ✅ Go testing with coverage reports
- ✅ GolangCI-Lint for code quality
- ✅ Security scanning with Gosec
- ✅ Docker build and push
- ✅ Branch-based deployments
- ✅ Post-build cleanup

**Setup**:
1. Copy `Jenkinsfile` to root of your repository
2. Install required Jenkins plugins:
   - Docker Pipeline
   - Cobertura Publisher (for coverage)
3. Configure Jenkins credentials:
   - Docker registry credentials
4. Set up Jenkins agents with Docker support
5. Adjust deployment commands for your infrastructure

## Common Configuration

### Environment Variables
All templates expect these environment variables (set in your CI/CD platform):

```bash
# Database (for integration tests if needed)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=test_db

# Application
APP_ENV=test
JWT_SECRET=test_secret
INTER_SERVICE_TOKEN=test_token
```

### Docker Registry
- **GitLab**: Uses built-in GitLab Container Registry
- **GitHub Actions**: Uses Docker Hub (configure secrets)
- **Jenkins**: Configure `DOCKER_REGISTRY` environment variable

### Deployment
Each template includes placeholder deployment steps. Customize these for your infrastructure:

- Kubernetes: `kubectl set image deployment/app app=image:tag`
- Docker Compose: `docker-compose up -d`
- AWS ECS: `aws ecs update-service`
- etc.

## Security Considerations

- Store secrets in CI/CD platform's secret management
- Use read-only tokens for Docker registries
- Implement proper RBAC for deployments
- Consider using separate service accounts for different environments

## Customization

Feel free to modify these templates based on your needs:
- Add more testing stages (integration tests, performance tests)
- Include database migrations in deployment
- Add notification steps (Slack, email)
- Implement blue-green or canary deployments
- Add monitoring and alerting

## Support

For issues with these templates, check:
- Go official documentation
- CI/CD platform documentation
- Tool documentation (golangci-lint, gosec, etc.)