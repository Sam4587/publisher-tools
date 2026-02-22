// service-manager.js - 服务管理器，监控和自动重启服务
const { spawn, exec } = require('child_process');
const fs = require('fs');
const path = require('path');

class ServiceManager {
  constructor() {
    this.services = new Map();
    this.logDir = path.join(__dirname, 'logs');
    this.ensureLogDir();
  }

  ensureLogDir() {
    if (!fs.existsSync(this.logDir)) {
      fs.mkdirSync(this.logDir, { recursive: true });
    }
  }

  // 启动服务
  startService(name, command, cwd, port) {
    if (this.services.has(name)) {
      console.log(`[${name}] 服务已在运行`);
      return;
    }

    console.log(`[${name}] 正在启动...`);
    
    const logFile = path.join(this.logDir, `${name}.log`);
    const logStream = fs.createWriteStream(logFile, { flags: 'a' });
    
    const proc = spawn(command, [], {
      cwd: cwd,
      shell: true,
      stdio: ['ignore', 'pipe', 'pipe']
    });

    proc.stdout.on('data', (data) => {
      const msg = `[${new Date().toISOString()}] ${data}`;
      logStream.write(msg);
      console.log(`[${name}] ${data.toString().trim()}`);
    });

    proc.stderr.on('data', (data) => {
      const msg = `[${new Date().toISOString()}] ERROR: ${data}`;
      logStream.write(msg);
      console.error(`[${name}] ERROR: ${data.toString().trim()}`);
    });

    proc.on('close', (code) => {
      console.log(`[${name}] 进程退出，代码: ${code}`);
      this.services.delete(name);
      
      // 自动重启
      if (code !== 0) {
        console.log(`[${name}] 5秒后自动重启...`);
        setTimeout(() => {
          this.startService(name, command, cwd, port);
        }, 5000);
      }
    });

    proc.on('error', (err) => {
      console.error(`[${name}] 启动失败:`, err);
    });

    this.services.set(name, {
      process: proc,
      command,
      cwd,
      port,
      startTime: Date.now()
    });

    console.log(`[${name}] 服务已启动 (PID: ${proc.pid})`);
  }

  // 停止服务
  stopService(name) {
    const service = this.services.get(name);
    if (!service) {
      console.log(`[${name}] 服务未运行`);
      return;
    }

    console.log(`[${name}] 正在停止...`);
    service.process.kill('SIGTERM');
    this.services.delete(name);
  }

  // 停止所有服务
  stopAll() {
    console.log('正在停止所有服务...');
    for (const [name] of this.services) {
      this.stopService(name);
    }
  }

  // 健康检查
  async healthCheck(name, port) {
    return new Promise((resolve) => {
      const http = require('http');
      const req = http.get(`http://localhost:${port}/api/health`, (res) => {
        resolve(res.statusCode === 200);
      });
      
      req.on('error', () => resolve(false));
      req.setTimeout(2000, () => {
        req.destroy();
        resolve(false);
      });
    });
  }

  // 监控服务状态
  async monitorServices() {
    console.log('\n=== 服务状态监控 ===');
    
    for (const [name, service] of this.services) {
      const uptime = Math.floor((Date.now() - service.startTime) / 1000);
      const isHealthy = await this.healthCheck(name, service.port);
      
      console.log(`[${name}]`);
      console.log(`  - PID: ${service.process.pid}`);
      console.log(`  - 端口: ${service.port}`);
      console.log(`  - 运行时间: ${uptime}秒`);
      console.log(`  - 健康状态: ${isHealthy ? '✅ 正常' : '❌ 异常'}`);
    }
    
    console.log('===================\n');
  }

  // 获取服务状态
  getStatus() {
    const status = {};
    for (const [name, service] of this.services) {
      status[name] = {
        pid: service.process.pid,
        port: service.port,
        uptime: Math.floor((Date.now() - service.startTime) / 1000)
      };
    }
    return status;
  }
}

// 主程序
async function main() {
  const manager = new ServiceManager();
  
  // 捕获退出信号
  process.on('SIGINT', () => {
    console.log('\n收到退出信号...');
    manager.stopAll();
    process.exit(0);
  });

  process.on('SIGTERM', () => {
    manager.stopAll();
    process.exit(0);
  });

  // 启动服务
  console.log('========================================');
  console.log('  Publisher Tools - 服务管理器');
  console.log('========================================\n');

  // 启动 Go 热点服务器
  manager.startService(
    'hotspot-server',
    'go run main.go',
    path.join(__dirname, 'hotspot-server'),
    8080
  );

  // 等待 3 秒
  await new Promise(resolve => setTimeout(resolve, 3000));

  // 启动 Node.js 后端
  manager.startService(
    'node-backend',
    'node simple-server.js',
    path.join(__dirname, 'server'),
    3001
  );

  // 等待 2 秒
  await new Promise(resolve => setTimeout(resolve, 2000));

  // 启动前端
  manager.startService(
    'frontend',
    'npm run dev',
    path.join(__dirname, 'publisher-web'),
    5173
  );

  // 定期监控
  setInterval(() => {
    manager.monitorServices();
  }, 30000); // 每 30 秒检查一次

  // 保持进程运行
  console.log('\n服务管理器运行中...');
  console.log('按 Ctrl+C 停止所有服务\n');
}

// 如果直接运行此脚本
if (require.main === module) {
  main().catch(console.error);
}

module.exports = ServiceManager;
